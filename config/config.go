package config

import (
	"os"
	"reflect"
	"strconv"
)

type Config struct {
	NamespaceMain  string `env:"NAMESPACE_MAIN" default:"gameservers"`
	RuntimeType    string `env:"RUNTIME_TYPE" default:"containerd"`
	ContainerdPath string `env:"CONTAINERD_PATH" default:"/run/containerd/containerd.sock"`
}

func ParseConfigFromEnv(cfg interface{}) {
	val := reflect.ValueOf(cfg).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		envKey := fieldType.Tag.Get("env")
		defaultValue := fieldType.Tag.Get("default")

		// Get the environment variable value
		value := os.Getenv(envKey)
		if value == "" {
			value = defaultValue
		}

		// Set the value on the struct field
		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int:
			if intValue, err := strconv.Atoi(value); err == nil {
				field.SetInt(int64(intValue))
			}
		}
	}
}
