package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

type ExportMetricsFile struct {
	ResourceMetrics []struct {
		Resource     common.Resource      `json:"resource"`
		ScopeMetrics []common.ScopeMetric `json:"scopeMetrics"`
		SchemaURL    string               `json:"schemaUrl"`
	} `json:"resourceMetrics"`
}

func ProcessMetricsFile() {
	log.Printf("📁 Creating output directory: %s", common.DebugDir)
	if err := os.MkdirAll(common.DebugDir, os.ModePerm); err != nil {
		log.Fatalf("❌ Failed to create metrics directory: %v", err)
	}

	log.Printf("📖 Reading metrics file: %s", common.InputFile)
	data, err := os.ReadFile(common.InputFile)
	if err != nil {
		log.Fatalf("❌ Failed to read metrics file: %v", err)
	}

	var export ExportMetricsFile
	log.Printf("🛠️ Parsing JSON data...")
	if err := json.Unmarshal(data, &export); err != nil {
		log.Fatalf("❌ Failed to unmarshal metrics JSON: %v", err)
	}

	count := 0
	for _, rm := range export.ResourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			count++
			fileName := filepath.Join(common.DebugDir, fmt.Sprintf("scopeMetrics_%03d.json", count))

			outputMetric := map[string]interface{}{
				"resource": rm.Resource,
				"scopeMetric": map[string]interface{}{
					"scope":     sm.Scope,
					"metrics":   sm.Metrics,
					"schemaUrl": sm.SchemaURL,
				},
			}

			outputJSON, err := json.MarshalIndent(outputMetric, "", "  ")
			if err != nil {
				log.Printf("❌ Failed to marshal scopeMetric: %v", err)
				continue
			}

			err = os.WriteFile(fileName, outputJSON, 0644)
			if err != nil {
				log.Printf("❌ Failed to write scopeMetric file %s: %v", fileName, err)
			} else {
				log.Printf("✅ Successfully wrote: %s", fileName)
			}
		}
	}

	log.Printf("📦 Wrote %d scopeMetric files", count)
}
