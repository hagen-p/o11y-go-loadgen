package common

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// configStruct defines how config.yaml is parsed
type configStruct struct {
	BaseCluster  string `yaml:"base_cluster"`
	BaseName     string `yaml:"base_name"`
	NoReplicas   int    `yaml:"no_replicas"`
	CollectorURL string `yaml:"collectorURL"`
	InputDir     string `yaml:"input_dir"`
	DebugDir     string `yaml:"debug_dir"`
	InputFile    string `yaml:"input_file"`
}

var replicasOverride int

// RegisterFlags allows other files to use --replicas
func RegisterFlags() {
	flag.IntVar(&replicasOverride, "replicas", 1, "Override the number of replicas in config.yaml")
}

// LoadConfig reads config.yaml, applies overrides, and validates fields
func LoadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("❌ Failed to read config file: %v", err)
	}

	var cfg configStruct
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("❌ Failed to parse config file: %v", err)
	}

	BaseClusterName = cfg.BaseCluster
	BaseNodeName = cfg.BaseName
	CollectorURL = cfg.CollectorURL
	InputDir = cfg.InputDir
	DebugDir = cfg.DebugDir

	if replicasOverride > 1 {
		NoReplicas = replicasOverride
		log.Printf("⚙️ Overriding replicas from CLI: %d", NoReplicas)
	} else {
		NoReplicas = cfg.NoReplicas
	}

	if NoReplicas <= 0 {
		log.Fatalf("❌ Invalid number of replicas: %d (must be > 0)", NoReplicas)
	}

	if expanded, err := ExpandPath(cfg.InputFile); err == nil {
		InputFile = expanded
	} else {
		InputFile = cfg.InputFile
	}
}

// Print selected fields from config for debug/info output
func PrintConfig(fields ...string) {
	log.Println("✅ Selected config values:")
	for _, field := range fields {
		switch field {
		case "BaseClusterName":
			log.Printf("  BaseClusterName: %s", BaseClusterName)
		case "BaseNodeName":
			log.Printf("  BaseNodeName:    %s", BaseNodeName)
		case "NoReplicas":
			log.Printf("  NoReplicas:      %d", NoReplicas)
		case "InputDir":
			log.Printf("  InputDir:        %s", InputDir)
		case "InputFile":
			log.Printf("  InputFile:       %s", InputFile)
		case "DebugDir":
			log.Printf("  DebugDir:        %s", DebugDir)
		case "CollectorURL":
			log.Printf("  CollectorURL:    %s", CollectorURL)
		default:
			log.Printf("  ⚠️ Unknown config field: %s", field)
		}
	}
}
