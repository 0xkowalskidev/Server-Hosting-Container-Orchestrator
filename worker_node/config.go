package workernode

type Config struct {
	NamespaceMain  string `env:"NAMESPACE_MAIN" default:"gameservers"`
	ContainerdPath string `env:"CONTAINERD_PATH" default:"/run/containerd/containerd.sock"`
}
