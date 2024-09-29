package controlnode

import (
	"fmt"
	"log"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
)

type Schedular struct {
	containerService *ContainerService
	nodeService      *NodeService
}

func NewSchedular(containerService *ContainerService, nodeService *NodeService) *Schedular {
	return &Schedular{
		containerService: containerService,
		nodeService:      nodeService,
	}
}

func (s *Schedular) ScheduleContainers() error {
	containers, err := s.containerService.GetContainers()
	if err != nil {
		// TODO: Handle me
	}

	nodes, err := s.nodeService.GetNodes()
	if err != nil {
		// TODO: Handle me
	}

	// TODO: Optimize me
	for _, container := range containers {
		if container.NodeID == "" {
			err := s.scheduleContainer(container, nodes)
			if err != nil {
				log.Printf("Error scheduling container %s: %v", container.ID, err)
				continue
			}
		}
	}

	return nil
}

func (s *Schedular) scheduleContainer(container models.Container, nodes []models.Node) error {
	for _, node := range nodes {
		if s.doesNodeHaveFreeResources(container, node) {
			container.NodeID = node.ID
			err := s.containerService.PutContainer(container)
			if err != nil {
				// TODO: Handle me
			}

			node.Containers = append(node.Containers, container)
			err = s.nodeService.PutNode(node)
			if err != nil {
				// TODO: Handle me
			}

			return nil
		}
	}

	return fmt.Errorf("failed to assign container %s", container.ID)
}

func (s *Schedular) doesNodeHaveFreeResources(container models.Container, node models.Node) bool {
	return true
}
