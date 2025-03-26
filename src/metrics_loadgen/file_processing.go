package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

// Process single JSON file
func processSingleFile() {
	if common.InputFile == "" {
		log.Println("❌ No input file specified in config.")
		return
	}

	processJSONFile(common.InputFile)
}

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

	dir := filepath.Dir(expandedPath)
	replacementsFile := filepath.Join(dir, "replacements.json")
	replacements := make(map[string]string)

	// Load previous replacements if the file exists
	if data, err := os.ReadFile(replacementsFile); err == nil {
		_ = json.Unmarshal(data, &replacements)
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

	nodeNameCounter := make(map[string]int)

	for clusterIndex := 0; clusterIndex < common.NoClusters; clusterIndex++ {
		metricsCopy := common.DeepCopyMetricsFile(metricsFile)
		clusterName := fmt.Sprintf("%s-%02d", common.BaseClusterName, clusterIndex)

		var resolvedNodeName string

		for i, attr := range metricsCopy.Resource.Attributes {
			key := attr.Key
			val := attr.Value.StringValue

			switch key {
			case "k8s.cluster.name":
				mappedKey := fmt.Sprintf("cluster:%s:%02d", val, clusterIndex)
				if replacement, ok := replacements[mappedKey]; ok {
					metricsCopy.Resource.Attributes[i].Value.StringValue = replacement
				} else {
					replacements[mappedKey] = clusterName
					metricsCopy.Resource.Attributes[i].Value.StringValue = clusterName
					log.Printf("🔄 Updating cluster name: %s -> %s", val, clusterName)
				}

			case "k8s.node.name":
				counterKey := fmt.Sprintf("%s-%02d", val, clusterIndex)
				count := nodeNameCounter[counterKey]
				nodeNameCounter[counterKey]++

				letter1 := 'A' + (count / 26)
				letter2 := 'A' + (count % 26)
				suffix := fmt.Sprintf("%c%c", letter1, letter2)
				newNodeName := fmt.Sprintf("%s-%s-%02d", val, suffix, clusterIndex)

				mappedKey := fmt.Sprintf("node:%s:%02d:%d", val, clusterIndex, count)
				if replacement, ok := replacements[mappedKey]; ok {
					metricsCopy.Resource.Attributes[i].Value.StringValue = replacement
				} else {
					replacements[mappedKey] = newNodeName
					metricsCopy.Resource.Attributes[i].Value.StringValue = newNodeName
					log.Printf("🔄 Updating node name: %s -> %s", val, newNodeName)
				}

				// Store resolved node name for syncing to host.name
				resolvedNodeName = metricsCopy.Resource.Attributes[i].Value.StringValue

			case "host.name":
				if resolvedNodeName != "" {
					metricsCopy.Resource.Attributes[i].Value.StringValue = resolvedNodeName
					log.Printf("🔄 Syncing host name to node name: %s", resolvedNodeName)
				}

			case "k8s.pod.uid":
				mappedKey := fmt.Sprintf("pod:%s:%02d", val, clusterIndex)
				if replacement, ok := replacements[mappedKey]; ok {
					metricsCopy.Resource.Attributes[i].Value.StringValue = replacement
				} else {
					newUID := fmt.Sprintf("uid-%s-%02d", val[:8], clusterIndex)
					replacements[mappedKey] = newUID
					metricsCopy.Resource.Attributes[i].Value.StringValue = newUID
					log.Printf("🔄 Updating pod UID: %s -> %s", val, newUID)
				}
			}
		}

		updateTimestamps(&metricsCopy)
		outputProcessedJSON(metricsCopy)
	}

	// Save updated replacements
	if jsonData, err := json.MarshalIndent(replacements, "", "  "); err == nil {
		_ = os.WriteFile(replacementsFile, jsonData, 0644)
	}
}
