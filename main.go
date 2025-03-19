package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Adjust struct to match JSON format
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
	AsInt             json.RawMessage `json:"asInt,omitempty"` // ✅ Handles both strings & numbers
	AsDouble          *float64        `json:"asDouble,omitempty"`
}

func main() {
	inputFile := "./metrics.json"
	outputDir := "./metric"

	// Create output directory
	log.Printf("📁 Creating output directory: %s", outputDir)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("❌ Failed to create metrics directory: %v", err)
	}

	// Read the entire file
	log.Printf("📖 Reading metrics file: %s", inputFile)
	data, err := os.ReadFile(inputFile)
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
			fileName := filepath.Join(outputDir, fmt.Sprintf("scopeMetrics_%d.json", counter))

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

	log.Printf("✅ Successfully split %d ScopeMetrics into separate files in .metric/\n", counter-1)
}
