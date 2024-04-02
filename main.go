package main

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func main() {
	if err := debugExample(); err != nil {
		log.Fatal(err)
	}
}

func debugExample() error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), "example")

	// Pull the busybox image
	image, err := client.Pull(ctx, "docker.io/itzg/minecraft-server:latest", containerd.WithPullUnpack)
	if err != nil {
		return err
	}
	log.Printf("Successfully pulled %s image\n", image.Name())

	// Create a new container using the busybox image
	container, err := client.NewContainer(
		ctx,
		"minecraft-server-container",
		containerd.WithImage(image),
		containerd.WithNewSnapshot("minecraft-server-snapshot", image),
		containerd.WithNewSpec(
			oci.WithHostNamespace(specs.NetworkNamespace),
			oci.WithImageConfig(image),
			oci.WithEnv([]string{"EULA=TRUE"}), // Set the EULA environment variable
			func(ctx context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {

				s.Mounts = append(s.Mounts, specs.Mount{
					Destination: "/etc/resolv.conf",
					Type:        "bind",
					Source:      "/home/kowalski/dev/container-orchestrator/resolv.conf",
					Options:     []string{"rbind", "ro"},
				})

				return nil
			},
		),
	)
	if err != nil {
		return err
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
	log.Printf("Successfully created container with ID %s and snapshot with ID minecraft-snapshot", container.ID())

	logPath := "/home/kowalski/dev/container-orchestrator/log.log"
	// Create a new task and start it
	task, err := container.NewTask(ctx, cio.LogFile(logPath))
	if err != nil {
		return err
	}
	defer task.Delete(ctx)

	if err := task.Start(ctx); err != nil {
		return err
	}

	// Wait for a signal to kill the task, for demonstration we wait for a fixed time
	time.Sleep(60 * time.Second) // Adjust the sleep time as necessary

	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return err
	}

	// Wait for the task to exit and get the exit status
	statusC, err := task.Wait(ctx)
	if err != nil {
		return err
	}
	status := <-statusC
	code, _, err := status.Result()
	if err != nil {
		return err
	}
	fmt.Printf("minecarft-container exited with status: %d\n", code)

	return nil
}
