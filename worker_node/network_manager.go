package workernode

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	cni "github.com/containerd/go-cni"
	"github.com/vishvananda/netns"
)

type NetworkManager struct {
	config    Config
	cninet    cni.CNI
	fileOps   utils.FileOpsInterface
	cmdRunner utils.CmdRunnerInterface
}

func NewNetworkManager(config Config, fileOps utils.FileOpsInterface, cmdRunner utils.CmdRunnerInterface) (*NetworkManager, error) {
	cninet, err := cni.New(
		cni.WithMinNetworkCount(2),
		cni.WithConfListFile("/etc/cni/net.d/10-mynet.conflist"), //TODO: Put these in config
		cni.WithPluginDir([]string{"/run/current-system/sw/bin"}),
	)
	if err != nil {
		return nil, err
	}

	if err := cninet.Load(cni.WithLoNetwork, cni.WithDefaultConf); err != nil {
		return nil, fmt.Errorf("failed to load CNI configurations: %v", err)
	}

	return &NetworkManager{config: config, cninet: cninet, fileOps: fileOps, cmdRunner: cmdRunner}, nil
}

func (nm *NetworkManager) SyncNetwork(desiredContainers []models.Container) {
	desiredMap := make(map[string]models.Container)
	for _, container := range desiredContainers {
		desiredMap[container.ID] = container
	}

	actualMap := nm.GetNetworkNamespaces()
	for containerID, container := range desiredMap {
		if !actualMap[containerID] {
			if err := nm.setupNetworking(container); err != nil {
				log.Printf("failed to set up networking for container %s: %v", containerID, err)
			}
		}
	}

	for containerID := range actualMap {
		if _, desired := desiredMap[containerID]; !desired {
			if err := nm.teardownNetworking(containerID); err != nil {
				log.Printf("failed to tear down networking for container %s: %v", containerID, err)
			}
		}
	}
}

func (nm *NetworkManager) setupNetworking(container models.Container) error {
	cmd := exec.Command("ip", "netns", "add", container.ID) // TODO: Put this in a cmd runner util if no better solution can be found
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("creating network namespace failed: %w", err)
	}

	netnsPath := fmt.Sprintf("/var/run/netns/%s", container.ID)

	var portMappings []cni.PortMapping
	for _, port := range container.Ports {
		portMappings = append(portMappings, cni.PortMapping{
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      port.Protocol,
		})
	}

	_, err = nm.cninet.Setup(context.Background(), container.ID, netnsPath, cni.WithCapabilityPortMap(portMappings))
	if err != nil {
		return fmt.Errorf("failed to set up networking with CNI: %v", err)
	}

	return nil
}

func (nm *NetworkManager) teardownNetworking(containerID string) error {
	netnsPath := fmt.Sprintf("/var/run/netns/%s", containerID)

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

	err = nm.cninet.Remove(context.Background(), containerID, netnsPath, cni.WithCapabilityPortMap(portMappings))
	if err != nil {
		return fmt.Errorf("failed to remove networking with CNI: %v", err)
	}

	err = netns.DeleteNamed(containerID) // TODO: Match the method used for creating netns
	if err != nil {
		return fmt.Errorf("Failed to delete network namespace: %v", err)
	}

	return nil
}

func (nm *NetworkManager) GetNetworkNamespaces() map[string]bool {
	netNsMap := make(map[string]bool)

	entries, err := nm.fileOps.ReadDir("/var/run/netns")
	if err != nil { // Err Likely means 404, which is fine as netns will be created automatically
		// TODO: Can I use errdefs here
		return netNsMap
	}

	for _, entry := range entries {
		netNsMap[entry.Name()] = true
	}

	return netNsMap
}

func (nm *NetworkManager) getContainerIP(containerID string) (string, error) {
	command := "nsenter"
	args := []string{"--net=" + "/var/run/netns/" + containerID, "ip", "addr", "show", "eth0"}

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

func (nm *NetworkManager) parseIpTablesOutput(output string) ([]cni.PortMapping, error) {
	var portMappings []cni.PortMapping
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
			portMapping := cni.PortMapping{
				HostPort:      int32(hostPort),
				ContainerPort: int32(containerPort),
				Protocol:      protocol,
			}
			portMappings = append(portMappings, portMapping)
		}
	}
	return portMappings, nil
}
