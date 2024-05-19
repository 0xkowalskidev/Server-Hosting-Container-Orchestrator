package controlnode

import (
	"0xKowalski1/container-orchestrator/models"
	"fmt"
	"log"
)

type Scheduler struct {
	etcdClient       *EtcdClient
	containerService *ContainerService
	nodeService      *NodeService
}

func NewScheduler(etcdClient *EtcdClient, containerService *ContainerService, nodeService *NodeService) *Scheduler {
	scheduler := &Scheduler{
		etcdClient:       etcdClient,
		containerService: containerService,
		nodeService:      nodeService,
	}

	// Run schedule once on start to handle any missed events while offline.
	scheduler.scheduleContainers()

	scheduler.etcdClient.Subscribe(func(event Event) {
		if event.Type == ContainerAdded {
			scheduler.scheduleContainers()
		}
	})

	return scheduler
}

func (s *Scheduler) scheduleContainers() {
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
		err := s.scheduleContainer(container, nodes)
		if err == nil {
			log.Printf("Container %s scheduled successfully", container.ID)
		} else {
			log.Printf("Error scheduling container %s: %v", container.ID, err)
		}
	}
}

func (s *Scheduler) scheduleContainer(container models.Container, nodes []models.Node) error {
	for _, node := range nodes {
		if !s.doesNodeHaveFreeResources(container, node) {
			log.Printf("Node %s does not have resources free to schedule container %s", node.ID, container.ID)
			continue
		}

		if !s.doesNodeHavePortsAvailable(container, node) {
			log.Printf("Node %s does not have ports free to schedule container %s", node.ID, container.ID)
			continue
		}

		err := s.nodeService.AssignContainerToNode(container.ID, node.ID)
		if err != nil {
			log.Printf("Failed to assign container %s to node %s: %v", container.ID, node.ID, err)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to assign container %s, nodes at capacity", container.ID)
}

func (s *Scheduler) doesNodeHaveFreeResources(container models.Container, node models.Node) bool {
	if node.MemoryLimit-node.MemoryUsed < container.MemoryLimit ||
		node.CpuLimit-node.CpuUsed < container.CpuLimit ||
		node.StorageLimit-node.StorageUsed < container.StorageLimit {
		return false
	}
	return true
}

func (s *Scheduler) doesNodeHavePortsAvailable(container models.Container, node models.Node) bool {
	usedPortsMap := make(map[int]bool)
	for _, c := range node.Containers {
		for _, port := range c.Ports {
			usedPortsMap[port.HostPort] = true
		}
	}

	for _, port := range container.Ports {
		if usedPortsMap[port.HostPort] {
			return false
		}
	}
	return true
}
