package models

type Volume struct {
	ID         string
	MountPoint string
	SizeLimit  int64 // Size in MB
}
