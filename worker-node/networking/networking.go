package networking

import (
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/utils"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"0xKowalski1/container-orchestrator/models"

	"github.com/containernetworking/cni/libcni"
)

type NetworkingManager struct {
	cfg       *config.Config
	cmdRunner utils.CmdRunnerInterface
}

func NewNetworkingManager(cfg *config.Config, cmdRunner utils.CmdRunnerInterface) *NetworkingManager {
	return &NetworkingManager{
		cfg:       cfg,
		cmdRunner: cmdRunner,
	}
}

func (nm *NetworkingManager) createNetworkNamespace(containerID string) error {
	err := nm.cmdRunner.RunCommand("ip", "netns", "add", containerID)
	if err != nil {
		return fmt.Errorf("creating network namespace failed: %w", err)
	}
	return nil
}

func (nm *NetworkingManager) deleteNetworkNamespace(containerID string) error {
	err := nm.cmdRunner.RunCommand("ip", "netns", "del", containerID)
	if err != nil {
		return fmt.Errorf("deleting network namespace failed: %w", err)
	}
	return nil
}

func (nm *NetworkingManager) ListNetworkNamespaces() ([]string, error) {
	output, err := nm.cmdRunner.RunCommandWithOutput("ip", "netns", "list")
	if err != nil {
		return nil, fmt.Errorf("running command failed: %w", err)
	}

	var namespaces []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			namespaces = append(namespaces, fields[0])
		}
	}
	return namespaces, nil
}

func (nm *NetworkingManager) getContainerIP(containerID string) (string, error) {
	command := "nsenter"
	args := []string{"--net=" + nm.cfg.NetworkNamespacePath + containerID, "ip", "addr", "show", "eth0"}

	output, err := nm.cmdRunner.RunCommandWithOutput(command, args...)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6 ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ip := strings.Split(fields[1], "/")[0] // IP address without CIDR notation
				return ip, nil
			}
		}
	}

	return "", fmt.Errorf("IP address not found")
}

func (nm *NetworkingManager) SetupContainerNetwork(containerID string, ports []models.Port) error {
	ctx := context.Background()
	cniConfig := libcni.CNIConfig{Path: []string{nm.cfg.CNIPath}}

	if err := nm.createNetworkNamespace(containerID); err != nil {
		return err
	}

	netConf, err := libcni.LoadConfList(nm.cfg.NetworkConfigPath, nm.cfg.NetworkConfigFileName)
	if err != nil {
		return fmt.Errorf("loading CNI configuration failed: %w", err)
	}

	var portMappings []map[string]interface{}
	for _, port := range ports {
		portMapping := map[string]interface{}{
			"hostPort":      port.HostPort,
			"containerPort": port.ContainerPort,
			"protocol":      port.Protocol,
		}
		portMappings = append(portMappings, portMapping)
	}

	runtimeConf := &libcni.RuntimeConf{
		ContainerID:    containerID,
		NetNS:          nm.cfg.NetworkNamespacePath + containerID,
		IfName:         "eth0",
		CapabilityArgs: map[string]interface{}{"portMappings": portMappings},
	}

	_, err = cniConfig.AddNetworkList(ctx, netConf, runtimeConf)
	if err != nil {
		return fmt.Errorf("setting up container network failed: %w", err)
	}

	return nil
}

func (nm *NetworkingManager) CleanupContainerNetwork(containerID string) error {
	ctx := context.Background()
	cniConfig := libcni.CNIConfig{Path: []string{nm.cfg.CNIPath}}

	netConf, err := libcni.LoadConfList(nm.cfg.NetworkConfigPath, nm.cfg.NetworkConfigFileName)
	if err != nil {
		return fmt.Errorf("loading CNI configuration failed: %w", err)
	}

	containerIP, err := nm.getContainerIP(containerID)
	if err != nil {
		return fmt.Errorf("Failed to get container ip: %v", err)
	}

	command := "bash"
	args := []string{"-c", fmt.Sprintf(`sudo iptables -t nat -S | grep %s`, containerIP)}
	output, err := nm.cmdRunner.RunCommandWithOutput(command, args...)
	if err != nil {
		return fmt.Errorf("failed to execute iptables command: %v", err)
	}

	portMappings, err := nm.parseIpTablesOutput(output)
	if err != nil {
		return fmt.Errorf("failed to parse iptables output: %v", err)
	}

	log.Println("Parsed Port Mappings:")
	for _, mapping := range portMappings {
		log.Printf("HostPort: %d, ContainerPort: %d, Protocol: %s\n",
			mapping["hostPort"], mapping["containerPort"], mapping["protocol"])
	}

	runtimeConf := &libcni.RuntimeConf{
		ContainerID:    containerID,
		NetNS:          nm.cfg.NetworkNamespacePath + containerID,
		IfName:         "eth0",
		CapabilityArgs: map[string]interface{}{"portMappings": portMappings},
	}

	log.Printf("Deleting CNI network for container %s", containerID)
	if err := cniConfig.DelNetworkList(ctx, netConf, runtimeConf); err != nil {
		log.Printf("Error cleaning up CNI network for container %s: %v", containerID, err)
		return fmt.Errorf("cleaning up container network failed: %w", err)
	}

	log.Printf("Deleting network namespace for container %s", containerID)
	if err := nm.deleteNetworkNamespace(containerID); err != nil {
		log.Printf("Error deleting network namespace for container %s: %v", containerID, err)
		return err
	}

	return nil
}

func (nm *NetworkingManager) parseIpTablesOutput(output string) ([]map[string]interface{}, error) {
	var portMappings []map[string]interface{}
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`-p (\w+) .* --dport (\d+) -j DNAT --to-destination \d+\.\d+\.\d+\.\d+:(\d+)`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 4 {
			protocol := matches[1]
			hostPort, err := strconv.Atoi(matches[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing hostPort: %v", err)
			}
			containerPort, err := strconv.Atoi(matches[3])
			if err != nil {
				return nil, fmt.Errorf("error parsing containerPort: %v", err)
			}
			portMapping := map[string]interface{}{
				"hostPort":      hostPort,
				"containerPort": containerPort,
				"protocol":      protocol,
			}
			portMappings = append(portMappings, portMapping)
		}
	}
	return portMappings, nil
}
