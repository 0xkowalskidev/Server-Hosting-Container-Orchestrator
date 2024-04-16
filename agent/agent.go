package agent

import (
	"log"
	"time"

	"0xKowalski1/container-orchestrator/api"
	"0xKowalski1/container-orchestrator/config"
	"0xKowalski1/container-orchestrator/models"
	"0xKowalski1/container-orchestrator/networking"
	"0xKowalski1/container-orchestrator/runtime"
	"0xKowalski1/container-orchestrator/storage"
)

type ApiResponse struct {
	Node struct {
		ID         string             `json:"ID"`
		Containers []models.Container `json:"Containers"`
	} `json:"node"`
}

type Agent struct {
	runtime    *runtime.ContainerdRuntime
	storage    *storage.StorageManager
	networking *networking.NetworkingManager
	cfg        *config.Config
}

func NewAgent(cfg *config.Config) *Agent {
	// start runtime
	runtime, err := runtime.NewContainerdRuntime(cfg)

	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	storage := storage.NewStorageManager(cfg)

	networking := networking.NewNetworkingManager(cfg)

	agent := &Agent{
		runtime:    runtime,
		storage:    storage,
		networking: networking,
		cfg:        cfg,
	}

	go agent.startLogApi()

	return agent
}

func (a *Agent) Start() {
	apiClient := api.NewApiWrapper(a.cfg.Namespace)

	nodeConfig := models.CreateNodeRequest{
		ID:           "node-1",
		MemoryLimit:  16,
		CpuLimit:     4,
		StorageLimit: 10,
		NodeIp:       a.cfg.NodeIp,
	}

	// Check if node exists
	_, err := apiClient.GetNode(nodeConfig.ID)
	if err != nil {
		// If it does not (or not authed, which will fail) try and Join Cluster
		_, err := apiClient.JoinCluster(nodeConfig)
		if err != nil {
			log.Printf("Error joining cluster: %v", err)
			panic(err)
		}
	}

	// If node already exists and it isnt use, then auth should catch it

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		node, err := apiClient.GetNode(nodeConfig.ID) //temp
		if err != nil {
			log.Printf("Error checking for nodes desired state: %v", err)
			continue
		}

		var desiredVolumes []models.Volume
		desiredContainers := node.Containers
		//Define wanted storage/containers/networking
		// Why am I even doing this?
		for _, desiredContainer := range node.Containers {
			newVolume := models.Volume{ID: desiredContainer.ID, SizeLimit: int64(desiredContainer.StorageLimit)}
			desiredVolumes = append(desiredVolumes, newVolume)
		}

		err = a.syncStorage(desiredVolumes)
		if err != nil {
			log.Printf("Error syncing storage: %v", err)
			continue
		}

		err = a.syncNetworking(desiredContainers)
		if err != nil {
			log.Printf("Error syncing network: %v", err)
			continue
		}

		err = a.syncContainers(node)
		if err != nil {
			log.Printf("Error syncing containers: %v", err)
			continue
		}

	}
}
