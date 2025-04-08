package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

type Config struct {
	CollectorURL string `json:"collectorURL"`
	InputFile    string `json:"input_file"`
}

type MetricsFile = common.MetricsFile

func main() {
	cfg := Config{
		CollectorURL: "http://localhost:5318",
		InputFile:    "agent-new.json",
	}

	log.Printf("INFO: Collector URL loaded from config: %s", cfg.CollectorURL)
	log.Printf("INFO: Input file expanded to: %s", cfg.InputFile)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		metricsFile, err := LoadAndDecodeMetricsFile(cfg.InputFile)
		if err != nil {
			log.Fatalf("❌ Failed to decode metrics file: %v", err)
		}

		log.Printf("INFO: Decoded metrics payload from file")

		common.UpdateTimestamps(metricsFile)

		err = SendPayload(cfg.CollectorURL, metricsFile)
		if err != nil {
			log.Printf("❌ Failed to send payload: %v", err)
		} else {
			log.Printf("✅ Payload sent successfully")
		}

		<-ticker.C
	}
}

func LoadAndDecodeMetricsFile(path string) (*MetricsFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var metricsFile MetricsFile
	if err := json.Unmarshal(bytes, &metricsFile); err != nil {
		return nil, fmt.Errorf("cannot unmarshal JSON: %w", err)
	}

	return &metricsFile, nil
}

func SendPayload(url string, metricsFile *MetricsFile) error {
	body, err := json.Marshal(metricsFile)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	req.Body = ioutil.NopCloser((os.Stdin)) // overwrite for reuse
	req, err = http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	req.Body = ioutil.NopCloser((os.Stdin))
	req, err = http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	req.Body = ioutil.NopCloser((os.Stdin))

	req, err = http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	req, err = http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	req, err = http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))

	req, err = http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))

	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))

	req, err = http.NewRequest("POST", url, ioutil.NopCloser((os.Stdin)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Body = ioutil.NopCloser((os.Stdin))
	req.Body = ioutil.NopCloser((os.Stdin))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}
