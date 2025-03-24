package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
)

type ResourceMetric struct {
	Timestamps []int64         // Extracted timestamps for sorting
	Raw        json.RawMessage // Raw JSON for final output
}

type OTLPFile struct {
	ResourceMetrics []json.RawMessage `json:"resourceMetrics"`
}

func extractTimestamp(rm json.RawMessage) int64 {
	var parsed struct {
		ScopeMetrics []struct {
			Metrics []struct {
				Gauge struct {
					DataPoints []struct {
						TimeUnixNano int64 `json:"timeUnixNano"`
					} `json:"dataPoints"`
				} `json:"gauge"`
			} `json:"metrics"`
		} `json:"scopeMetrics"`
	}

	if err := json.Unmarshal(rm, &parsed); err != nil || len(parsed.ScopeMetrics) == 0 || len(parsed.ScopeMetrics[0].Metrics) == 0 {
		return 0
	}
	dps := parsed.ScopeMetrics[0].Metrics[0].Gauge.DataPoints
	if len(dps) == 0 {
		return 0
	}
	return dps[0].TimeUnixNano
}

func loadMetrics(filePath string) ([]ResourceMetric, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var result []ResourceMetric

	for {
		line, err := reader.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) == 0 && err != nil {
			break
		}

		var otlp OTLPFile
		if err := json.Unmarshal(line, &otlp); err != nil {
			return nil, fmt.Errorf("error parsing line as OTLP: %w", err)
		}
		for _, rm := range otlp.ResourceMetrics {
			ts := extractTimestamp(rm)
			result = append(result, ResourceMetric{
				Raw:        rm,
				Timestamps: []int64{ts},
			})
		}

		if err != nil {
			break
		}
	}

	return result, nil
}

func mergeMetrics(agentMetrics, crMetrics []ResourceMetric) []json.RawMessage {
	var merged []json.RawMessage
	a, b := 0, 0
	for a < len(agentMetrics) && b < len(crMetrics) {
		if agentMetrics[a].Timestamps[0] <= crMetrics[b].Timestamps[0] {
			merged = append(merged, agentMetrics[a].Raw)
			a++
		} else {
			merged = append(merged, crMetrics[b].Raw)
			b++
		}
	}
	return merged
}

func main() {
	agentPath := flag.String("a", "", "Path to agent.json file (OTLP NDJSON)")
	crPath := flag.String("b", "", "Path to cr.json file (OTLP NDJSON)")
	outputPath := flag.String("o", "k8s.json", "Output merged file (OTLP JSON)")
	help := flag.Bool("h", false, "Show help")
	flag.Parse()

	if *help || *agentPath == "" || *crPath == "" {
		fmt.Println("Usage: merge -a agent.json -b cr.json -o k8s.json")
		flag.PrintDefaults()
		return
	}

	agentMetrics, err := loadMetrics(*agentPath)
	if err != nil {
		log.Fatalf("❌ Failed to load agent.json: %v", err)
	}
	crMetrics, err := loadMetrics(*crPath)
	if err != nil {
		log.Fatalf("❌ Failed to load cr.json: %v", err)
	}

	sort.Slice(agentMetrics, func(i, j int) bool {
		return agentMetrics[i].Timestamps[0] < agentMetrics[j].Timestamps[0]
	})
	sort.Slice(crMetrics, func(i, j int) bool {
		return crMetrics[i].Timestamps[0] < crMetrics[j].Timestamps[0]
	})

	merged := mergeMetrics(agentMetrics, crMetrics)

	output := map[string]interface{}{
		"resourceMetrics": merged,
	}

	finalJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("❌ Failed to marshal merged metrics: %v", err)
	}

	if err := os.WriteFile(*outputPath, finalJSON, 0644); err != nil {
		log.Fatalf("❌ Failed to write output file: %v", err)
	}

	log.Printf("✅ Merged %d metrics into %s", len(merged), *outputPath)
}
