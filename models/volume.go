package models

// Volume is used on the worker node in the storage manager
type Volume struct {
	ID         string
	MountPoint string
	SizeLimit  int64 // Size in GB
}
