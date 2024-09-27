package controlnode

type Config struct {
	Namespace string `env:"NAMESPACE" default:"gameservers"` // Namespace for etcd/containerd
}
