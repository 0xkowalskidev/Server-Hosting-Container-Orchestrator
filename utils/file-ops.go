package utils

import "os"

type FileOpsInterface interface {
	Stat(path string) (os.FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error
	ReadDir(dirname string) ([]os.DirEntry, error)
	Remove(name string) error
}

type FileOps struct{}

func (f *FileOps) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (f *FileOps) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *FileOps) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (f *FileOps) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}

func (f *FileOps) Remove(name string) error {
	return os.Remove(name)
}
