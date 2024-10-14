package controlnode

import (
	"testing"
	"time"

	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/models"
	"github.com/0xKowalskiDev/Server-Hosting-Container-Orchestrator/utils"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var validContainer = models.Container{
	ID: "valid-container",
}

func setup(t *testing.T) ContainerService {
	var config Config
	utils.ParseConfigFromEnv(&config)

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"}, // TODO: Get these from config
		DialTimeout: 5 * time.Second,
	})

	containerService := NewContainerService(config, etcdClient)

	require.NoError(t, err)
	require.NotNil(t, etcdClient)

	return *containerService
}

func TestCreateContainer_Valid(t *testing.T) {
	containerService := setup(t)

	err := containerService.PutContainer(validContainer)

	require.NoError(t, err)
}
