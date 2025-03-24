package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

// Update k8s.cluster.name for multiple clusters
func updateClusterNames(metricsFile *common.MetricsFile) {
	for clusterIndex := 0; clusterIndex < common.NoClusters; clusterIndex++ {
		clusterName := fmt.Sprintf("%s-%d", common.BaseClusterName, clusterIndex)

		for i, attr := range metricsFile.Resource.Attributes {
			if attr.Key == "k8s.cluster.name" {
				log.Printf("🔄 Updating cluster name: %s -> %s", attr.Value.StringValue, clusterName)
				metricsFile.Resource.Attributes[i].Value.StringValue = clusterName
			}
		}
	}
}

// Update timestamps while keeping original differences
func updateTimestamps(metricsFile *common.MetricsFile) {
	currentTime := time.Now().UnixNano()

	for _, metric := range metricsFile.ScopeMetric.Metrics {
		for i := range metric.Gauge.DataPoints {
			startStr := metric.Gauge.DataPoints[i].StartTimeUnixNano
			endStr := metric.Gauge.DataPoints[i].TimeUnixNano

			// Fallback time difference: 5 nanoseconds
			const defaultDiff = 5

			var (
				originalStartTime time.Duration
				originalTime      time.Duration
				err1, err2        error
			)

			if startStr != "" {
				originalStartTime, err1 = time.ParseDuration(startStr + "ns")
			} else {
				err1 = fmt.Errorf("StartTimeUnixNano is empty")
			}

			if endStr != "" {
				originalTime, err2 = time.ParseDuration(endStr + "ns")
			} else {
				err2 = fmt.Errorf("TimeUnixNano is empty")
			}

			timeDiff := time.Duration(defaultDiff)
			if err1 == nil && err2 == nil {
				timeDiff = originalTime - originalStartTime
			} else {
				log.Printf("⚠️ Using default 5ns diff due to timestamp parse error: %v, %v", err1, err2)
			}

			// Apply the time difference
			metric.Gauge.DataPoints[i].StartTimeUnixNano = fmt.Sprintf("%d", currentTime)
			metric.Gauge.DataPoints[i].TimeUnixNano = fmt.Sprintf("%d", currentTime+int64(timeDiff))
		}
	}
}

// Send the processed JSON to OTLP receiver via HTTP
/* func old_outputProcessedJSON(metricsFile common.MetricsFile) {
	outputJSON, err := json.Marshal(metricsFile)
	if err != nil {
		log.Printf("❌ Failed to marshal updated JSON: %v", err)
		return
	}

	// OTLP HTTP metrics endpoint
	otlpURL := "http://localhost:5318/v1/metrics"

	// Create HTTP POST request
	req, err := http.NewRequest("POST", otlpURL, bytes.NewBuffer(outputJSON))
	if err != nil {
		log.Printf("❌ Failed to create HTTP request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Dump and pause before sending
	req, err = common.DumpAndPauseRequest(req, outputJSON)
	if err != nil {
		log.Fatal(err)
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send OTLP data: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("✅ Successfully sent OTLP metrics to %s (status: %s)", otlpURL, resp.Status)
	} else {
		log.Printf("⚠️ Unexpected response from OTLP receiver: %s", resp.Status)
	}
} */

func outputProcessedJSON(metricsFile common.MetricsFile) {
	// ⚠️ Manually construct an OTLP ExportMetricsServiceRequest
	otlpRequest := &collector.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricpb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: common.ToOTLPAttributes(metricsFile.Resource.Attributes),
				},
				ScopeMetrics: []*metricpb.ScopeMetrics{
					{
						Scope: common.ToOTLPScope(metricsFile.ScopeMetric.Scope),
						//Metrics: []*metricpb.Metric{common.ToOTLPMetrics(metricsFile.ScopeMetric.Metrics)...},
						Metrics: common.ToOTLPMetrics(metricsFile.ScopeMetric.Metrics),
					},
				},
			},
		},
	}

	// Marshal to OTLP/JSON format
	outputJSON, err := protojson.Marshal(otlpRequest)
	if err != nil {
		log.Printf("❌ Failed to marshal OTLP JSON: %v", err)
		return
	}

	otlpURL := "http://localhost:5318/v1/metrics"

	req, err := http.NewRequest("POST", otlpURL, bytes.NewBuffer(outputJSON))
	if err != nil {
		log.Printf("❌ Failed to create HTTP request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Debug out put for the request body and heacders
	/* fmt.Println("DEBUG: Raw marshaled JSON:")
	fmt.Println(string(outputJSON))
	req, err = common.DumpAndPauseRequest(req, outputJSON)
	if err != nil {
		log.Fatal(err)
	} */

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send OTLP data: %v", err)
		return
	}
	defer resp.Body.Close()

	//   debug output for  http response body
	/* 	body, err := io.ReadAll(resp.Body)
	   	if err != nil {
	   		log.Printf("⚠️ Could not read response body: %v", err)
	   	} else if len(body) > 0 {
	   		log.Println("📩 Collector response body:")
	   		log.Println(string(body))
	   	}
	*/
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("✅ Successfully sent OTLP metrics to %s (status: %s)", otlpURL, resp.Status)
	} else {
		log.Printf("⚠️ Unexpected response from OTLP receiver: %s", resp.Status)
	}
}
