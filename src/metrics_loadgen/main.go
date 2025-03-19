package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

// Global configuration variables
var (
	BaseClusterName string
	NoClusters      int
	AccessToken     string
	RumToken        string
	ApiToken        string
	InputDir        string
	InputFile       string
	OutputDir       string
)

// Struct for parsing config.yaml
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

// Structs matching JSON format
type MetricsFile struct {
	ResourceMetrics []ResourceMetric `json:"resourceMetrics"`
}

type ResourceMetric struct {
	Resource     Resource      `json:"resource"`
	ScopeMetrics []ScopeMetric `json:"scopeMetrics"`
	SchemaURL    string        `json:"schemaUrl"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value struct {
		StringValue string `json:"stringValue"`
	} `json:"value"`
}

type ScopeMetric struct {
	Scope     Scope    `json:"scope"`
	Metrics   []Metric `json:"metrics"`
	SchemaURL string   `json:"schemaUrl"`
}

type Scope struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Metric struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit,omitempty"`
	Gauge       struct {
		DataPoints []DataPoint `json:"dataPoints"`
	} `json:"gauge,omitempty"`
}

type DataPoint struct {
	StartTimeUnixNano string          `json:"startTimeUnixNano"`
	TimeUnixNano      string          `json:"timeUnixNano"`
	AsInt             json.RawMessage `json:"asInt,omitempty"`
	AsDouble          *float64        `json:"asDouble,omitempty"`
}

func main() {
	// Define command-line flags
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Display usage information")
	flag.Parse()

	// Show help message if -h is passed
	if *helpFlag {
		fmt.Println("Usage: metrics_loadgen [options]")
		fmt.Println("Options:")
		fmt.Println("  --config=<path>  Specify the configuration file (default: config.yaml)")
		fmt.Println("  -h               Display this help message")
		os.Exit(0)
	}

	// Load configuration
	loadConfig(*configPath)

	log.Printf("📂 Monitoring directory: %s", InputDir)

	// Signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("🛑 Stopping JSON processing...")
		os.Exit(0)
	}()

	// Start looping through JSON files
	for {
		processFiles()
		time.Sleep(2 * time.Second) // Avoid high CPU usage
	}
}

func loadConfig(configPath string) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatalf("❌ Failed to parse config file: %v", err)
	}

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

func processFiles() {
	expandedPath, err := expandPath(InputDir)
	if err != nil {
		log.Printf("❌ Failed to expand input directory path: %v", err)
		return
	}

	files, err := os.ReadDir(expandedPath)
	if err != nil {
		log.Printf("❌ Failed to read input directory: %v", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(expandedPath, file.Name())
			processJSONFile(filePath)
		}
	}
}

func processJSONFile(filePath string) {
	expandedPath, err := expandPath(filePath)
	if err != nil {
		log.Printf("❌ Failed to expand file path: %v", err)
		return
	}

	log.Printf("📖 Processing file: %s", expandedPath)

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		log.Printf("❌ Failed to read file: %s, error: %v", expandedPath, err)
		return
	}

	var metricsFile MetricsFile
	if err := json.Unmarshal(data, &metricsFile); err != nil {
		log.Printf("❌ Failed to unmarshal JSON: %v", err)
		return
	}

	// Update the `k8s.cluster.name` field
	for clusterIndex := 0; clusterIndex < NoClusters; clusterIndex++ {
		clusterName := fmt.Sprintf("%s-%d", BaseClusterName, clusterIndex)

		for _, resourceMetric := range metricsFile.ResourceMetrics {
			for i, attr := range resourceMetric.Resource.Attributes {
				if attr.Key == "k8s.cluster.name" {
					log.Printf("🔄 Updating cluster name: %s -> %s", attr.Value.StringValue, clusterName)
					resourceMetric.Resource.Attributes[i].Value.StringValue = clusterName
				}
			}
		}
	}

	// Call output function instead of saving to a file
	outputProcessedJSON(metricsFile)
}

func outputProcessedJSON(metricsFile MetricsFile) {
	outputJSON, err := json.MarshalIndent(metricsFile, "", "  ")
	if err != nil {
		log.Printf("❌ Failed to marshal updated JSON: %v", err)
		return
	}
	log.Println("📝 Processed JSON Output:")
	fmt.Println(string(outputJSON))
}

func expandPath(path string) (string, error) {
	if len(path) > 0 {
		if path[0] == '~' {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, path[1:]), nil
		}
		return filepath.Abs(path)
	}
	return path, nil
}
