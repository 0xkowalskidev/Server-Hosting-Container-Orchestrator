package workernode

type Config struct {
	NodeID         string `env:"NODE_ID"`
	ControlNodeURI string `env:"CONTROL_NODE_URI"`
	ContainerdPath string `env:"CONTAINERD_PATH" default:"/run/containerd/containerd.sock"`
}
