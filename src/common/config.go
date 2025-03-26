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
	CollectorURL    string `yaml:"collectorURL"`
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
	CollectorURL = config.CollectorURL

	// Expand paths (expandPath can also live in utils.go)
	InputDir, _ = ExpandPath(config.InputDir)
	InputFile, _ = ExpandPath(config.InputFile)
	OutputDir, _ = ExpandPath(config.OutputDir)

	PrintConfig("InputFile", "CollectorURL")
}

func PrintConfig(fields ...string) {
	log.Println("✅ Selected config values:")
	for _, field := range fields {
		switch field {
		case "BaseClusterName":
			log.Printf("  BaseClusterName: %s", BaseClusterName)
		case "NoClusters":
			log.Printf("  NoClusters:      %d", NoClusters)
		case "AccessToken":
			log.Printf("  AccessToken:     %s", AccessToken)
		case "RumToken":
			log.Printf("  RumToken:        %s", RumToken)
		case "ApiToken":
			log.Printf("  ApiToken:        %s", ApiToken)
		case "InputDir":
			log.Printf("  InputDir:        %s", InputDir)
		case "InputFile":
			log.Printf("  InputFile:       %s", InputFile)
		case "OutputDir":
			log.Printf("  OutputDir:       %s", OutputDir)
		case "CollectorURL":
			log.Printf("  CollectorURL:    %s", CollectorURL)
		default:
			log.Printf("  ⚠️ Unknown config field: %s", field)
		}
	}
}
