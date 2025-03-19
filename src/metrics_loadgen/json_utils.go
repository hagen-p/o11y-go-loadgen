package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Update k8s.cluster.name for multiple clusters
func updateClusterNames(metricsFile *MetricsFile) {
	for clusterIndex := 0; clusterIndex < NoClusters; clusterIndex++ {
		clusterName := fmt.Sprintf("%s-%d", BaseClusterName, clusterIndex)

		for i, attr := range metricsFile.Resource.Attributes {
			if attr.Key == "k8s.cluster.name" {
				log.Printf("🔄 Updating cluster name: %s -> %s", attr.Value.StringValue, clusterName)
				metricsFile.Resource.Attributes[i].Value.StringValue = clusterName
			}
		}
	}
}

// Update timestamps while keeping original differences
func updateTimestamps(metricsFile *MetricsFile) {
	currentTime := time.Now().UnixNano()
	for _, metric := range metricsFile.ScopeMetric.Metrics {
		for i := range metric.Gauge.DataPoints {
			originalStartTime, err1 := time.ParseDuration(metric.Gauge.DataPoints[i].StartTimeUnixNano + "ns")
			originalTime, err2 := time.ParseDuration(metric.Gauge.DataPoints[i].TimeUnixNano + "ns")

			if err1 != nil || err2 != nil {
				log.Printf("❌ Failed to parse timestamps: %v, %v", err1, err2)
				continue
			}

			timeDifference := originalTime - originalStartTime

			// Apply the same time difference
			metric.Gauge.DataPoints[i].StartTimeUnixNano = fmt.Sprintf("%d", currentTime)
			metric.Gauge.DataPoints[i].TimeUnixNano = fmt.Sprintf("%d", currentTime+int64(timeDifference))
		}
	}
}

// Output JSON to console (or modify to send to OTLP later)
func outputProcessedJSON(metricsFile MetricsFile) {
	outputJSON, err := json.MarshalIndent(metricsFile, "", "  ")
	if err != nil {
		log.Printf("❌ Failed to marshal updated JSON: %v", err)
		return
	}
	log.Println("📝 Processed JSON Output:")
	fmt.Println(string(outputJSON))
}
