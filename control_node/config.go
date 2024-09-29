package controlnode

type Config struct {
	EtcdNamespace string `env:"ETCD_NAMESPACE" default:"gameservers"` // Namespace prefix for etcd
}
