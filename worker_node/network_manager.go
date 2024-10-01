package workernode

import (
	"context"
	"fmt"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	cni "github.com/containerd/go-cni"
	"github.com/vishvananda/netns"
	"log"
)

type NetworkManager struct {
	config  Config
	cninet  cni.CNI
	fileOps utils.FileOpsInterface
}

func NewNetworkManager(config Config, fileOps utils.FileOpsInterface) (*NetworkManager, error) {
	cninet, err := cni.New(
		cni.WithConfListFile("/etc/cni/net.d/10-mynet.conflist"), //TODO: Put these in config
		cni.WithPluginDir([]string{"/run/current-system/sw/bin"}),
	)
	if err != nil {
		return nil, err
	}

	if err := cninet.Load(cni.WithDefaultConf); err != nil {
		return nil, fmt.Errorf("failed to load CNI configurations: %v", err)
	}

	return &NetworkManager{config: config, cninet: cninet, fileOps: fileOps}, nil
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
	newNS, err := netns.NewNamed(container.ID)
	if err != nil {
		return fmt.Errorf("failed to create new network namespace: %v", err)
	}
	defer newNS.Close()

	netnsPath := fmt.Sprintf("/var/run/netns/%s", container.ID)

	var portMappings []cni.PortMapping // TODO: Is this needed?
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

	err := nm.cninet.Remove(context.Background(), containerID, netnsPath)
	if err != nil {
		return fmt.Errorf("failed to remove networking with CNI: %v", err)
	}

	err = netns.DeleteNamed(containerID)
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
