package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

// sendToCollector sends the OTLP JSON payload to the configured collector URL.
func sendToCollector(payload []byte) {
	otlpURL := common.CollectorURL + "/v1/metrics"

	req, err := http.NewRequest("POST", otlpURL, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("❌ Failed to create HTTP request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send OTLP data: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("⚠️ Non-success response: %s - %s", resp.Status, string(body))
	} else {
		log.Printf("✅ Payload sent successfully: %s", resp.Status)
	}
}

// writePayloadToFile writes the payload to a timestamped OTLP JSON file for offline analysis.
func writePayloadToFile(payload []byte) {
	dir := "../outbox"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("❌ Failed to create output directory %s: %v", dir, err)
		return
	}

	timestamp := time.Now().Format("2006-01-02T15-04-05.000000000")
	filename := fmt.Sprintf("payload-%s.json", timestamp)
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, payload, 0644); err != nil {
		log.Printf("❌ Failed to write payload file: %v", err)
	} else {
		log.Printf("📤 Payload written to: %s", path)
	}
}
