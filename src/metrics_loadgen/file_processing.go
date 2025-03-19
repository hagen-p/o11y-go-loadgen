package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// Process JSON files in the input directory
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

// Process a single JSON file
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

	// Modify JSON data
	updateClusterNames(&metricsFile)
	updateTimestamps(&metricsFile)

	// Send or save the updated JSON
	outputProcessedJSON(metricsFile)
}
