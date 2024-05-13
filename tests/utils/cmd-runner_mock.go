package utils_test

import (
	"github.com/stretchr/testify/mock"
)

type MockCmdRunner struct {
	mock.Mock
}

func (c *MockCmdRunner) RunCommand(name string, args ...string) error {
	fullArgs := append([]interface{}{name}, stringSliceToInterfaceSlice(args)...)
	return c.Called(fullArgs...).Error(0)
}

func (c *MockCmdRunner) RunCommandWithOutput(name string, args ...string) (string, error) {
	fullArgs := append([]interface{}{name}, stringSliceToInterfaceSlice(args)...)
	argsMock := c.Called(fullArgs...)
	return argsMock.String(0), argsMock.Error(1)
}

func stringSliceToInterfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, s := range strings {
		result[i] = s
	}
	return result
}
