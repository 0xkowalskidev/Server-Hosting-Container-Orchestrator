package utils

import "os/exec"

type CmdRunnerInterface interface {
	RunCommand(name string, args ...string) error
}

type CmdRunner struct{}

func (c *CmdRunner) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
