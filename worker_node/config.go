package workernode

type Config struct {
	NodeID              string `env:"NODE_ID"`
	ControlNodeURI      string `env:"CONTROL_NODE_URI"`
	ContainerdNamespace string `env:"CONTAINERD_NAMESPACE" default:"gameservers"` // Namespace for containerd
	ContainerdPath      string `env:"CONTAINERD_PATH" default:"/run/containerd/containerd.sock"`
	LogsPath            string `env:"LOGS_PATH"`   // Absolute logs path
	MountsPath          string `env:"MOUNTS_PATH"` // Absolute mounts path
	SftpPort            string `env:"SFTP_PORT"`
	SftpKeyPath         string `env:"SFTP_KEY_PATH"` // Absolute sftp priv key path
}
