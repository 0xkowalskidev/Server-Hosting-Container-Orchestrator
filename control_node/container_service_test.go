package controlnode

import (
	"context"
	"fmt"
	"log"
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

func setup(t *testing.T) (ContainerService, clientv3.Client) {
	var config Config
	utils.ParseConfigFromEnv(&config)

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"}, // TODO: Get these from config
		DialTimeout: 5 * time.Second,
	})

	require.NoError(t, err)

	containerService := NewContainerService(config, etcdClient)

	teardown(*etcdClient, containerService.config)

	return *containerService, *etcdClient
}

func teardown(etcdClient clientv3.Client, config Config) {
	_, err := etcdClient.Delete(context.Background(), fmt.Sprintf("/%s", config.EtcdNamespace), clientv3.WithPrefix())

	if err != nil {
		log.Println("Failed to teardown etcd")
	}
}

func TestCreateContainer_Valid(t *testing.T) {
	containerService, etcdClient := setup(t)
	defer teardown(etcdClient, containerService.config)

	err := containerService.PutContainer(validContainer)

	require.NoError(t, err)
}
