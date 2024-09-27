package workernode

type Config struct {
	ContainerdPath string `env:"CONTAINERD_PATH" default:"/run/containerd/containerd.sock"`
}
