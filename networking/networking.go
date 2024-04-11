package networking

import (
	"0xKowalski1/container-orchestrator/config"
	"context"
	"fmt"
	"log"
	"os/exec"
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

	if err := nm.cleanupIPRulesByIP(containerID); err != nil {
		return err
	}

	runtimeConf := &libcni.RuntimeConf{
		ContainerID: containerID,
		NetNS:       nm.cfg.NetworkNamespacePath + containerID,
		IfName:      "eth0",
	}

	if err := cniConfig.DelNetworkList(ctx, netConf, runtimeConf); err != nil {
		return fmt.Errorf("cleaning up container network failed: %w", err)
	}

	if err := nm.deleteNetworkNamespace(containerID); err != nil {
		return err
	}

	return nil
}

func (nm *NetworkingManager) cleanupIPRulesByIP(containerID string) error {
	containerIP, err := nm.getContainerIP(containerID)
	if err != nil {
		return fmt.Errorf("error getting container IP: %v", err)
	}
	log.Printf("Container IP to cleanup: %s", containerIP)

	tables := []string{"nat", "filter"} // Focus on relevant tables

	for _, table := range tables {
		// List all rules in the table
		cmd := exec.Command("iptables", "-t", table, "-S")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("listing iptables rules failed: %w", err)
		}

		// Iterate through each listed rule
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, containerIP) {
				// Rule found; construct deletion command
				parts := strings.Fields(line)
				if len(parts) > 2 {
					// The first part is "-A" (add), change it to "-D" (delete) for the deletion command
					parts[0] = "-D"
					delCmd := exec.Command("iptables", append([]string{"-t", table}, parts...)...)
					log.Printf("Executing: %v", delCmd) // Log the command for debugging

					// Attempt to execute the deletion command
					if err := delCmd.Run(); err != nil {
						log.Printf("Failed to delete rule: %v", err) // Log failure
					} else {
						log.Printf("Successfully deleted rule: %s", line) // Log success
					}
				}
			}
		}
	}

	return nil
}
