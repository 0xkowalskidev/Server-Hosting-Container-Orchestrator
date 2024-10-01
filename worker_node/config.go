package workernode

type Config struct {
	NodeID              string `env:"NODE_ID"`
	ControlNodeURI      string `env:"CONTROL_NODE_URI"`
	ContainerdNamespace string `env:"CONTAINERD_NAMESPACE" default:"gameservers"` // Namespace for containerd
	ContainerdPath      string `env:"CONTAINERD_PATH" default:"/run/containerd/containerd.sock"`
	LogsPath            string `env:"LOGS_PATH"` // Absolute logs path
}
