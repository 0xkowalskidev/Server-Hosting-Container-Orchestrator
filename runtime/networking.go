package runtime

import (
	"0xKowalski1/container-orchestrator/models"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containernetworking/cni/libcni"
)

var networkNamespacePathPrefix = "/var/run/netns/" // TAKE ME FROM CONFIG

// This cant be the best way of doing this
// Create a network namespace
func createNetworkNamespace(containerID string) error {
	cmd := exec.Command("ip", "netns", "add", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating network namespace failed: %w", err)
	}
	return nil
}

// Delete a network namespace
func deleteNetworkNamespace(containerID string) error {
	cmd := exec.Command("ip", "netns", "del", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deleting network namespace failed: %w", err)
	}
	return nil
}

func getContainerIP(containerID string) (string, error) {
	// Construct the command to list IP addresses within the network namespace
	cmd := exec.Command("nsenter", "--net="+networkNamespacePathPrefix+containerID, "ip", "addr", "show", "eth0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	// Parse the command output to find the IP address
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

// Cleanup IP rules for a specific container IP
func cleanupIPRulesByIP(containerID string) error {
	containerIP, err := getContainerIP(containerID)
	if err != nil {
		return fmt.Errorf("Error getting container IP: %v", err)
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
					delCmd := exec.Command("iptables", "-t", table)
					delCmd.Args = append(delCmd.Args, parts...)
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

func setupContainerNetwork(ctx context.Context, container containerd.Container, ports []models.Port) error {
	// Load CNI configuration
	cniConfig := libcni.CNIConfig{
		Path: []string{"/run/current-system/sw/bin"}, // TAKE ME FROM CONFIG
	}

	if err := createNetworkNamespace(container.ID()); err != nil {
		log.Printf("Failed to create network namespace: %v", err)
		return err
	}

	// Assuming your CNI network configuration file is named '10-my-network.conflist'
	netConf, err := libcni.LoadConfList("/etc/cni/net.d", "mynet") // FROM CONFIG
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

	// Prepare CNI runtime configuration, here you specify port mappings
	runtimeConf := &libcni.RuntimeConf{
		ContainerID: container.ID(),
		NetNS:       networkNamespacePathPrefix + container.ID(),
		IfName:      "eth0",
		CapabilityArgs: map[string]interface{}{
			"portMappings": portMappings,
		},
	}

	// Setup the network
	_, err = cniConfig.AddNetworkList(ctx, netConf, runtimeConf)
	if err != nil {
		return fmt.Errorf("setting up container network failed: %w", err)
	}

	return nil
}

func cleanupContainerNetwork(ctx context.Context, containerID string) error {
	cniConfig := libcni.CNIConfig{
		Path: []string{"/run/current-system/sw/bin"}, // TAKE ME FROM CONFIG
	}

	netConf, err := libcni.LoadConfList("/etc/cni/net.d", "mynet") // FROM CONFIG
	if err != nil {
		return fmt.Errorf("loading CNI configuration failed: %w", err)
	}

	runtimeConf := &libcni.RuntimeConf{
		ContainerID: containerID,
		NetNS:       networkNamespacePathPrefix + containerID,
		IfName:      "eth0",
	}

	err = cleanupIPRulesByIP(containerID)

	if err != nil {
		return err
	}

	// Delete the network
	err = cniConfig.DelNetworkList(ctx, netConf, runtimeConf)
	if err != nil {
		return fmt.Errorf("cleaning up container network failed: %w", err)
	}

	if err := deleteNetworkNamespace(containerID); err != nil {
		return err
	}

	return nil
}
