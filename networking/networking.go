package networking

import (
	"0xKowalski1/container-orchestrator/config"
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"0xKowalski1/container-orchestrator/models"

	"github.com/containernetworking/cni/libcni"
)

type NetworkingManager struct {
	cfg *config.Config
}

func NewNetworkingManager(cfg *config.Config) *NetworkingManager {
	return &NetworkingManager{
		cfg: cfg,
	}
}

func (nm *NetworkingManager) createNetworkNamespace(containerID string) error {
	cmd := exec.Command("ip", "netns", "add", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating network namespace failed: %w", err)
	}
	return nil
}

func (nm *NetworkingManager) deleteNetworkNamespace(containerID string) error {
	cmd := exec.Command("ip", "netns", "del", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deleting network namespace failed: %w", err)
	}
	return nil
}

func (nm *NetworkingManager) ListNetworkNamespaces() ([]string, error) {
	cmd := exec.Command("ip", "netns", "list")

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("running command failed: %w", err)
	}

	var namespaces []string
	lines := strings.Split(stdoutBuf.String(), "\n")
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
	cmd := exec.Command("nsenter", "--net="+nm.cfg.NetworkNamespacePath+containerID, "ip", "addr", "show", "eth0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	for _, line := range strings.Split(string(output), "\n") {
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

	cmd := exec.Command("bash", "-c", fmt.Sprintf(`sudo iptables -t nat -S | grep %s`, containerIP))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute iptables command: %v", err)
	}

	portMappings, err := nm.parseIpTablesOutput(string(output))
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
	log.Printf("%s", output)
	var portMappings []map[string]interface{}
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`-p (\w+) .* --dport (\d+) -j DNAT --to-destination \d+\.\d+\.\d+\.\d+:(\d+)`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 4 { // Ensure we are matching the entire pattern including container port
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
