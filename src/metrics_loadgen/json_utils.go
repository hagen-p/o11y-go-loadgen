package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	collector "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

func updateClusterNames(metricsFile *common.MetricsFile) {
	for clusterIndex := 0; clusterIndex < common.NoReplicas; clusterIndex++ {
		clusterName := fmt.Sprintf("%s-%d", common.BaseClusterName, clusterIndex)
		for _, rm := range metricsFile.ResourceMetrics {
			for i, attr := range rm.Resource.Attributes {
				if attr.Key == "k8s.cluster.name" {
					log.Printf("🔄 Updating cluster name: %s -> %s", attr.Value.StringValue, clusterName)
					rm.Resource.Attributes[i].Value.StringValue = clusterName
				}
			}
		}
	}
}

func updateTimestamps(metricsFile *common.MetricsFile) {
	currentTime := time.Now().UnixNano()

	for _, rm := range metricsFile.ResourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			for _, metric := range sm.Metrics {
				if metric.Gauge != nil {
					for i := range metric.Gauge.DataPoints {
						startStr := metric.Gauge.DataPoints[i].StartTimeUnixNano
						endStr := metric.Gauge.DataPoints[i].TimeUnixNano

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

						metric.Gauge.DataPoints[i].StartTimeUnixNano = fmt.Sprintf("%d", currentTime)
						metric.Gauge.DataPoints[i].TimeUnixNano = fmt.Sprintf("%d", currentTime+int64(timeDiff))
					}
				}
			}
		}
	}
}

func outputProcessedJSON(metricsFile common.MetricsFile) {
	var otlpResourceMetrics []*metricpb.ResourceMetrics

	for _, rm := range metricsFile.ResourceMetrics {
		var scopeMetricsList []*metricpb.ScopeMetrics
		for _, sm := range rm.ScopeMetrics {
			scopeMetricsList = append(scopeMetricsList, &metricpb.ScopeMetrics{
				Scope:     common.ToOTLPScope(sm.Scope),
				SchemaUrl: sm.SchemaURL,
				Metrics:   common.ToOTLPMetrics(sm.Metrics),
			})
		}

		otlpResourceMetrics = append(otlpResourceMetrics, &metricpb.ResourceMetrics{
			Resource: &resourcepb.Resource{
				Attributes: common.ToOTLPAttributes(rm.Resource.Attributes),
			},
			ScopeMetrics: scopeMetricsList,
			SchemaUrl:    rm.SchemaUrl,
		})
	}

	otlpRequest := &collector.ExportMetricsServiceRequest{
		ResourceMetrics: otlpResourceMetrics,
	}

	outputJSON, err := protojson.Marshal(otlpRequest)
	if err != nil {
		log.Printf("❌ Failed to marshal OTLP JSON: %v", err)
		return
	}

	otlpURL := common.CollectorURL + "/v1/metrics"

	req, err := http.NewRequest("POST", otlpURL, bytes.NewBuffer(outputJSON))
	if err != nil {
		log.Printf("❌ Failed to create HTTP request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if common.DebugEnabled {
		_ = os.WriteFile("console.out", outputJSON, 0644)
	}
	/*if common.DebugEnabled { // DEBUG output
		fmt.Println("DEBUG: Raw marshaled JSON:")
		fmt.Println(string(outputJSON))
		req, err = common.DumpAndPauseRequest(req, outputJSON)
		if err != nil {
			log.Fatal(err)
		}
	}
	*/
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send OTLP data: %v", err)
		return
	}
	defer resp.Body.Close()

	/*
		if common.DebugEnabled { // Debug output
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("⚠️ Could not read response body: %v", err)
			} else if len(body) > 0 {
				log.Println("📩 Collector response body:")
				log.Println(string(body))
			}
		}
	*/
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("✅ Successfully sent OTLP metrics to %s (status: %s)", otlpURL, resp.Status)
	} else {
		log.Printf("⚠️ Unexpected response from OTLP receiver: %s", resp.Status)
	}
}
