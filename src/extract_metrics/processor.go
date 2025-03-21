package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

func ProcessMetricsFile() {
	log.Printf("📁 Creating output directory: %s", common.OutputDir)
	if err := os.MkdirAll(common.OutputDir, os.ModePerm); err != nil {
		log.Fatalf("❌ Failed to create metrics directory: %v", err)
	}

	log.Printf("📖 Reading metrics file: %s", common.InputFile)
	data, err := os.ReadFile(common.InputFile)
	if err != nil {
		log.Fatalf("❌ Failed to read metrics file: %v", err)
	}

	log.Printf("🔍 JSON starts with: %s", string(data)[:500])

	var metricsFile common.MetricsFile
	log.Printf("🛠️ Parsing JSON data...")
	if err := json.Unmarshal(data, &metricsFile); err != nil {
		log.Fatalf("❌ Failed to unmarshal metrics JSON: %v", err)
	}

	// Since this is a single Resource + ScopeMetric, treat it as one entry
	log.Printf("📌 Processing single Resource and ScopeMetric")

	fileName := filepath.Join(common.OutputDir, "scopeMetrics_1.json")

	outputMetric := map[string]interface{}{
		"resource": metricsFile.Resource,
		"scopeMetric": map[string]interface{}{
			"scope":     metricsFile.ScopeMetric.Scope,
			"metrics":   metricsFile.ScopeMetric.Metrics,
			"schemaUrl": metricsFile.ScopeMetric.SchemaURL,
		},
	}

	outputJSON, err := json.MarshalIndent(outputMetric, "", "  ")
	if err != nil {
		log.Printf("❌ Failed to marshal scopeMetric: %v", err)
		return
	}

	err = os.WriteFile(fileName, outputJSON, 0644)
	if err != nil {
		log.Printf("❌ Failed to write scopeMetric file %s: %v", fileName, err)
	} else {
		log.Printf("✅ Successfully wrote scopeMetric file: %s", fileName)
	}
}
