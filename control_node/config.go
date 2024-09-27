package controlnode

type Config struct {
	Namespace string `env:"NAMESPACE" default:"development"` // Etcd Namespace
}
