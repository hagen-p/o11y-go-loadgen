package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

// Process JSON files in the input directory
func processFiles() {
	expandedPath, err := common.ExpandPath(common.InputDir)
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

// Process a single JSON file
func processJSONFile(filePath string) {
	expandedPath, err := common.ExpandPath(filePath)
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

	var metricsFile common.MetricsFile
	if err := json.Unmarshal(data, &metricsFile); err != nil {
		log.Printf("❌ Failed to unmarshal JSON: %v", err)
		return
	}

	// Modify JSON data
	updateClusterNames(&metricsFile)
	updateTimestamps(&metricsFile)

	// Send or save the updated JSON
	outputProcessedJSON(metricsFile)
}
