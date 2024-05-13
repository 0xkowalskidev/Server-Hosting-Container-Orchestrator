package utils

import (
	"bytes"
	"os/exec"
)

type CmdRunnerInterface interface {
	RunCommand(name string, args ...string) error
	RunCommandWithOutput(name string, args ...string) (string, error)
}

type CmdRunner struct{}

func (c *CmdRunner) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func (c *CmdRunner) RunCommandWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return stdoutBuf.String(), nil
}
