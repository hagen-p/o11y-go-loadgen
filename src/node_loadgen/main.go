package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

const (
	defaultInputFile = "agent.json"
	interval         = 10 * time.Second // ⏱️ New interval between rounds
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Display usage information")

	flag.BoolVar(&common.DebugEnabled, "d", false, "Enable debug output")
	flag.BoolVar(&common.InfoEnabled, "I", false, "Enable info-level logs to stdout")
	common.RegisterFlags()
	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage: node_loadgen [options]")
		fmt.Println("Options:")
		fmt.Println("  --config=<path>  Specify the configuration file (default: config.yaml)")
		fmt.Println("  --replicas=<n>   Override number of replicas from config")
		fmt.Println("  -d               Enable debug logs")
		fmt.Println("  -I               Enable info logs to stdout")
		fmt.Println("  -h               Display this help message")
		os.Exit(0)
	}

	common.InitLogging()
	common.LoadConfig(*configPath)

	if common.CollectorURL == "" {
		log.Println("❌ No Collector URL specified in config.")
		return
	}

	if common.InfoEnabled {
		log.Println("INFO: Collector URL loaded from config:", common.CollectorURL)
	}

	if common.InputFile == "" {
		log.Println("❌ No input file specified in config.")
		return
	}

	expandedPath, err := common.ExpandPath(common.InputFile)
	if err != nil {
		log.Printf("❌ Failed to expand file path: %v", err)
		return
	}
	if common.InfoEnabled {
		log.Println("INFO: Input file expanded to:", expandedPath)
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		log.Fatalf("Error opening input file '%s': %v", expandedPath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(bufio.NewReader(file))

	if common.InfoEnabled {
		log.Printf("INFO: Starting load generation loop. Sending to %s every %v", common.CollectorURL, interval)
	}
	debugOut := &bytes.Buffer{}

	for decoder.More() {
		var payload common.MetricsFile
		if err := decoder.Decode(&payload); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}

		if common.InfoEnabled {
			log.Println("INFO: Decoded metrics payload from file")
		}

		if common.DebugEnabled {
			for _, rm := range payload.ResourceMetrics {
				for _, attr := range rm.Resource.Attributes {
					line := fmt.Sprintf("Resource attribute: %s = %s\n", attr.Key, attr.Value.StringValue)
					debugOut.WriteString(line)
					log.Print(line[:len(line)-1])
				}
			}
			if err := os.WriteFile("../debug.out", debugOut.Bytes(), 0644); err != nil {
				log.Printf("❌ Failed to write ../debug.out: %v", err)
			} else {
				log.Println("🐞 Wrote resource attributes to ../debug.out")
			}
		}

		// Extract base host.id (string-based, e.g., i-abcdef1234567890)
		var baseHostID string
		for _, rm := range payload.ResourceMetrics {
			for _, attr := range rm.Resource.Attributes {
				if attr.Key == "host.id" {
					baseHostID = attr.Value.StringValue
					break
				}
			}
			if baseHostID != "" {
				break
			}
		}

		if common.NoReplicas > 1 {
			for i := 1; i <= common.NoReplicas; i++ {
				replica := common.DeepCopyMetricsFile(payload)
				nodeName := fmt.Sprintf("%s-%02d", common.BaseNodeName, i)

				for rmIdx := range replica.ResourceMetrics {
					attrs := &replica.ResourceMetrics[rmIdx].Resource.Attributes
					for j := range *attrs {
						switch (*attrs)[j].Key {
						case "node.name", "host.name":
							(*attrs)[j].Value.StringValue = nodeName
						case "host.id":
							if baseHostID != "" {
								(*attrs)[j].Value.StringValue = fmt.Sprintf("%s-%02d", baseHostID, i)
							}
						}
					}
				}

				if common.InfoEnabled {
					log.Printf("INFO: Generated replica %d with node.name = %s", i, nodeName)
				}

				common.UpdateTimestamps(&replica)

				buf, err := json.Marshal(replica)
				if err != nil {
					log.Printf("Failed to marshal replica payload: %v", err)
					continue
				}

				if common.DebugEnabled {
					log.Println("🐞 Debug mode: printing to file")
					go writePayloadToFile(buf)
				} else {
					go sendToCollector(buf)
				}
			}
		} else {
			if baseHostID != "" {
				for rmIdx := range payload.ResourceMetrics {
					attrs := &payload.ResourceMetrics[rmIdx].Resource.Attributes
					for j := range *attrs {
						if (*attrs)[j].Key == "host.id" {
							(*attrs)[j].Value.StringValue = fmt.Sprintf("%s-01", baseHostID)
						}
					}
				}
			}

			common.UpdateTimestamps(&payload)
			buf, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Failed to marshal payload: %v", err)
				continue
			}
			if common.DebugEnabled {
				log.Println("🐞 Debug mode: printing to file")
				go writePayloadToFile(buf)
			} else {
				go sendToCollector(buf)
			}
		}

		time.Sleep(interval) // ✅ sleep once per round, not per replica
	}
}
