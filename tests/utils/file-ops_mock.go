package utils_test

import (
	"os"
	"time"

	"github.com/stretchr/testify/mock"
)

type FakeDirEntry struct {
	name  string
	isDir bool
}

func NewFakeDirEntry(name string, isDir bool) FakeDirEntry {
	return FakeDirEntry{name: name, isDir: isDir}
}
func (f FakeDirEntry) Name() string               { return f.name }
func (f FakeDirEntry) IsDir() bool                { return f.isDir }
func (f FakeDirEntry) Type() os.FileMode          { return 0 }
func (f FakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }

type FakeFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func NewFakeFileInfo(name string, size int64, isDir bool) FakeFileInfo {
	return FakeFileInfo{name: name, size: size, isDir: isDir}
}
func (f FakeFileInfo) Name() string       { return f.name }
func (f FakeFileInfo) Size() int64        { return f.size }
func (f FakeFileInfo) Mode() os.FileMode  { return 0 }
func (f FakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f FakeFileInfo) IsDir() bool        { return f.isDir }
func (f FakeFileInfo) Sys() interface{}   { return nil }

// MockFileOps is a mock type for the FileOpsInterface
type MockFileOps struct {
	mock.Mock
}

func (m *MockFileOps) Stat(path string) (os.FileInfo, error) {
	args := m.Called(path)
	return args.Get(0).(os.FileInfo), args.Error(1)
}

func (m *MockFileOps) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *MockFileOps) RemoveAll(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileOps) ReadDir(dirname string) ([]os.DirEntry, error) {
	args := m.Called(dirname)
	return args.Get(0).([]os.DirEntry), args.Error(1)
}

func (m *MockFileOps) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}
