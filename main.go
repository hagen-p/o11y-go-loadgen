package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	OutputDir       string `yaml:"output_dir"`
}

// Structs to match JSON format
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
	// Load configuration
	loadConfig("config.yaml")

	// Create output directory
	log.Printf("📁 Creating output directory: %s", OutputDir)
	if err := os.MkdirAll(OutputDir, os.ModePerm); err != nil {
		log.Fatalf("❌ Failed to create metrics directory: %v", err)
	}

	// Read the metrics file
	metricsFilePath := filepath.Join(InputDir, "metrics.json")
	log.Printf("📖 Reading metrics file: %s", metricsFilePath)
	data, err := os.ReadFile(metricsFilePath)
	if err != nil {
		log.Fatalf("❌ Failed to read metrics file: %v", err)
	}

	// Debug JSON structure
	log.Printf("🔍 JSON starts with: %s", string(data)[:500]) // Print first 500 chars

	// Parse JSON as an object
	var metricsFile MetricsFile
	log.Printf("🛠️ Parsing JSON data...")
	if err := json.Unmarshal(data, &metricsFile); err != nil {
		log.Fatalf("❌ Failed to unmarshal metrics JSON: %v", err)
	}

	// Debugging number of ResourceMetrics
	log.Printf("📌 Found %d ResourceMetrics", len(metricsFile.ResourceMetrics))

	counter := 1
	for _, resourceMetric := range metricsFile.ResourceMetrics {
		log.Printf("📌 ResourceMetric has %d ScopeMetrics", len(resourceMetric.ScopeMetrics))

		for _, scopeMetric := range resourceMetric.ScopeMetrics {
			fileName := filepath.Join(OutputDir, fmt.Sprintf("scopeMetrics_%d.json", counter))

			log.Printf("🔹 Writing ScopeMetric #%d: Scope Name: %s", counter, scopeMetric.Scope.Name)

			outputMetric := map[string]interface{}{
				"resource": resourceMetric.Resource,
				"scopeMetric": map[string]interface{}{
					"scope":     scopeMetric.Scope,
					"metrics":   scopeMetric.Metrics,
					"schemaUrl": scopeMetric.SchemaURL,
				},
			}

			// Convert to JSON
			outputJSON, err := json.MarshalIndent(outputMetric, "", "  ")
			if err != nil {
				log.Printf("❌ Failed to marshal scopeMetric %d: %v", counter, err)
				continue
			}

			// Write to a separate file
			err = os.WriteFile(fileName, outputJSON, 0644)
			if err != nil {
				log.Printf("❌ Failed to write scopeMetric file %s: %v", fileName, err)
			} else {
				log.Printf("✅ Successfully wrote scopeMetric file: %s", fileName)
			}

			counter++
		}
	}

	log.Printf("✅ Successfully split %d ScopeMetrics into separate files in %s\n", counter-1, OutputDir)
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
	InputDir = config.InputDir
	OutputDir = config.OutputDir

	log.Printf("✅ Loaded config: BaseClusterName=%s, NoClusters=%d, InputDir=%s, OutputDir=%s",
		BaseClusterName, NoClusters, InputDir, OutputDir)
}
