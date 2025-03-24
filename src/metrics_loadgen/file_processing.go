package main

import (
	"encoding/json"
	"fmt"
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

	// 🔹 This is where you define the original metricsFile
	var metricsFile common.MetricsFile
	if err := json.Unmarshal(data, &metricsFile); err != nil {
		log.Printf("❌ Failed to unmarshal JSON: %v", err)
		return
	}

	// 🔁 Now loop over clusters
	for clusterIndex := 0; clusterIndex < common.NoClusters; clusterIndex++ {
		// 🔹 Deep copy the original for this cluster
		metricsCopy := common.DeepCopyMetricsFile(metricsFile)

		/* 		// 🔁 Update cluster name
		   		clusterName := fmt.Sprintf("%s-%d", common.BaseClusterName, clusterIndex)
		   		for i, attr := range metricsCopy.Resource.Attributes {
		   			if attr.Key == "k8s.cluster.name" {
		   				log.Printf("🔄 Updating cluster name: %s -> %s", attr.Value.StringValue, clusterName)
		   				metricsCopy.Resource.Attributes[i].Value.StringValue = clusterName
		   			}
		   		} */
		// Map to track how many times we've seen a node name per cluster
		nodeNameCounter := make(map[string]int)

		// 🔁 Update cluster and node name
		clusterName := fmt.Sprintf("%s-%02d", common.BaseClusterName, clusterIndex)

		for i, attr := range metricsCopy.Resource.Attributes {
			if attr.Key == "k8s.cluster.name" {
				log.Printf("🔄 Updating cluster name: %s -> %s", attr.Value.StringValue, clusterName)
				metricsCopy.Resource.Attributes[i].Value.StringValue = clusterName
			}
			if attr.Key == "k8s.node.name" {
				originalNodeName := attr.Value.StringValue

				// Track how many times we've renamed this base node in this cluster
				counterKey := fmt.Sprintf("%s-%02d", originalNodeName, clusterIndex)
				count := nodeNameCounter[counterKey]
				nodeNameCounter[counterKey]++

				// Convert count to AA, AB, AC, ..., ZZ
				letter1 := 'A' + (count / 26)
				letter2 := 'A' + (count % 26)
				suffix := fmt.Sprintf("%c%c", letter1, letter2)

				newNodeName := fmt.Sprintf("%s-%s-%02d", originalNodeName, suffix, clusterIndex)

				log.Printf("🔄 Updating node name: %s -> %s", originalNodeName, newNodeName)
				metricsCopy.Resource.Attributes[i].Value.StringValue = newNodeName
			}
		}

		// 🕒 Adjust timestamps
		updateTimestamps(&metricsCopy)

		// 🚀 Send it
		outputProcessedJSON(metricsCopy)
	}
}
