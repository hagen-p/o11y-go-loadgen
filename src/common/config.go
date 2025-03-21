package common

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the structure of the YAML configuration
type Config struct {
	BaseClusterName string `yaml:"base_cluster_name"`
	NoClusters      int    `yaml:"no_clusters"`
	AccessToken     string `yaml:"access_token"`
	RumToken        string `yaml:"rum_token"`
	ApiToken        string `yaml:"api_token"`
	InputDir        string `yaml:"input_dir"`
	InputFile       string `yaml:"input_file"`
	OutputDir       string `yaml:"output_dir"`
}

// LoadConfig reads and parses the config file, updating shared globals
func LoadConfig(configPath string) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatalf("❌ Failed to parse config file: %v", err)
	}

	// Assign global variables (from globals.go)
	BaseClusterName = config.BaseClusterName
	NoClusters = config.NoClusters
	AccessToken = config.AccessToken
	RumToken = config.RumToken
	ApiToken = config.ApiToken

	// Expand paths (expandPath can also live in utils.go)
	InputDir, _ = ExpandPath(config.InputDir)
	InputFile, _ = ExpandPath(config.InputFile)
	OutputDir, _ = ExpandPath(config.OutputDir)

	log.Printf("✅ Loaded config: BaseClusterName=%s, NoClusters=%d, InputDir=%s, OutputDir=%s",
		BaseClusterName, NoClusters, InputDir, OutputDir)
}
