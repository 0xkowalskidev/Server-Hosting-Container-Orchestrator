package utils_test

import (
	"github.com/stretchr/testify/mock"
)

type MockCmdRunner struct {
	mock.Mock
}

func (c *MockCmdRunner) RunCommand(name string, args ...string) error {
	// Append the command name and the args to form a single slice
	fullArgs := append([]interface{}{name}, stringSliceToInterfaceSlice(args)...)
	return c.Called(fullArgs...).Error(0)
}

// Helper function to convert slice of strings to slice of interfaces
func stringSliceToInterfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, s := range strings {
		result[i] = s
	}
	return result
}
