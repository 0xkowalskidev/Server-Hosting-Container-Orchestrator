package main

import (
	"fmt"
	"os"

	"0xKowalski1/container-orchestrator/agent"
	"0xKowalski1/container-orchestrator/config"
)

func main() {

	cfgPath := "config.json"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Join Cluster

	agent := agent.NewAgent(cfg)

	// start agent
	agent.Start()

}
