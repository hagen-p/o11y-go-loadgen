package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Load configuration from a YAML file
func loadConfig(configPath string) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatalf("❌ Failed to parse config file: %v", err)
	}

	// Assign global variables from the parsed config
	BaseClusterName = config.BaseClusterName
	NoClusters = config.NoClusters
	AccessToken = config.AccessToken
	RumToken = config.RumToken
	ApiToken = config.ApiToken

	// Expand paths
	InputDir, _ = expandPath(config.InputDir)
	InputFile, _ = expandPath(config.InputFile)
	OutputDir, _ = expandPath(config.OutputDir)

	log.Printf("✅ Loaded config: BaseClusterName=%s, NoClusters=%d, InputDir=%s, OutputDir=%s",
		BaseClusterName, NoClusters, InputDir, OutputDir)
}
