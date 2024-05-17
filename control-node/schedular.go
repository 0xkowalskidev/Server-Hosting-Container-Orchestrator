package controlnode

import (
	"log"
)

type Schedular struct {
	etcdClient       *EtcdClient
	containerService *ContainerService
	nodeService      *NodeService
}

func NewSchedular(etcdClient *EtcdClient, containerService *ContainerService, nodeService *NodeService) *Schedular {
	schedular := &Schedular{
		etcdClient:       etcdClient,
		containerService: containerService,
		nodeService:      nodeService,
	}

	// Run schedule once on start, incase there where missed unscheduled events while offline
	schedular.scheduleContainers()

	schedular.etcdClient.Subscribe(func(event Event) {
		switch event.Type {
		case ContainerAdded:
			schedular.scheduleContainers()
		}
	})

	return schedular
}

// scheduleContainers checks and updates the _statemanager to schedule containers to nodes.
func (s *Schedular) scheduleContainers() {
	unscheduledContainers, err := s.containerService.GetUnscheduledContainers()
	if err != nil {
		log.Printf("Error fetching unscheduled containers: %v", err)
		return
	}

	nodes, err := s.nodeService.GetNodes()
	if err != nil {
		log.Printf("Error fetching nodes: %v", err)
		return
	}

	log.Printf("Starting container scheduling...")
	for _, container := range unscheduledContainers {
		log.Printf("Attempting to schedule container %s...", container.ID)
		assigned := false
		for _, node := range nodes {
			if node.MemoryLimit-node.MemoryUsed >= container.MemoryLimit && node.CpuLimit-node.CpuUsed >= container.CpuLimit && node.StorageLimit-node.StorageUsed >= container.StorageLimit {
				// Check ports
				portsFree := true
				usedPortsMap := make(map[int]bool)
				for _, container := range node.Containers {
					for _, port := range container.Ports {
						usedPortsMap[port.HostPort] = true
					}
				}

				for _, port := range container.Ports {
					if usedPortsMap[port.HostPort] {
						portsFree = false
						continue
					}
				}

				if !portsFree {
					continue
				}

				err := s.nodeService.AssignContainerToNode(container.ID, node.ID)
				if err != nil {
					log.Printf("Failed to assign container %s to node %s: %v", container.ID, node.ID, err)
					continue
				}
				assigned = true
				break
			}
		}
		if !assigned {
			log.Printf("Failed to assign container %s, all nodes are at capacity", container.ID)
		}
	}
}
