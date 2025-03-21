package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func processMetricsFile() {
	log.Printf("📁 Creating output directory: %s", OutputDir)
	if err := os.MkdirAll(OutputDir, os.ModePerm); err != nil {
		log.Fatalf("❌ Failed to create metrics directory: %v", err)
	}

	log.Printf("📖 Reading metrics file: %s", InputFile)
	data, err := os.ReadFile(InputFile)
	if err != nil {
		log.Fatalf("❌ Failed to read metrics file: %v", err)
	}

	log.Printf("🔍 JSON starts with: %s", string(data)[:500])

	var metricsFile MetricsFile
	log.Printf("🛠️ Parsing JSON data...")
	if err := json.Unmarshal(data, &metricsFile); err != nil {
		log.Fatalf("❌ Failed to unmarshal metrics JSON: %v", err)
	}

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

			outputJSON, err := json.MarshalIndent(outputMetric, "", "  ")
			if err != nil {
				log.Printf("❌ Failed to marshal scopeMetric %d: %v", counter, err)
				continue
			}

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
