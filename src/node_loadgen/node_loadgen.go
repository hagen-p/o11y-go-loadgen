package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

const (
	defaultInputFile = "agent.json"
	interval         = 5 * time.Second // interval between sends
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Display usage information")

	flag.BoolVar(&common.DebugEnabled, "d", false, "Enable debug output")
	flag.BoolVar(&common.InfoEnabled, "I", false, "Enable info-level logs to stdout")

	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage: node_loadgen [options]")
		fmt.Println("Options:")
		fmt.Println("  --config=<path>  Specify the configuration file (default: config.yaml)")
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

		updateTimestamps(&payload)

		if common.InfoEnabled {
			log.Println("INFO: Updated timestamps in payload")
		}

		buf, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload: %v", err)
			continue
		}

		if common.InfoEnabled {
			log.Println("INFO: Sending payload to collector")
		}
		if common.DebugEnabled {
			log.Println("🐞 Debug mode: printing to file")
			go writePayloadToFile(buf)
		} else {
			go sendToCollector(buf)
		}

		time.Sleep(interval)
	}
}

func updateTimestamps(metricsFile *common.MetricsFile) {
	currentTime := time.Now().UnixNano()
	const defaultDiff = int64(5)

	for _, rm := range metricsFile.ResourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			for _, metric := range sm.Metrics {
				if metric.Gauge != nil {
					if len(metric.Gauge.DataPoints) == 0 {
						log.Printf("⚠️ Empty gauge: %s", metric.Name)
					}
					for i := range metric.Gauge.DataPoints {
						log.Printf("⏱️ %s: start=%s, time=%s",
							metric.Name,
							metric.Gauge.DataPoints[i].StartTimeUnixNano,
							metric.Gauge.DataPoints[i].TimeUnixNano,
						)
						updateGenericDatapointTimestamps(&metric.Gauge.DataPoints[i], currentTime, defaultDiff)
					}
				}
				if metric.Sum != nil {
					for i := range metric.Sum.DataPoints {
						log.Printf("⏱️ %s: start=%s, time=%s",
							metric.Name,
							metric.Sum.DataPoints[i].StartTimeUnixNano,
							metric.Sum.DataPoints[i].TimeUnixNano,
						)
						updateGenericDatapointTimestamps(&metric.Sum.DataPoints[i], currentTime, defaultDiff)
					}
				}
				if metric.Histogram != nil {
					for i := range metric.Histogram.DataPoints {
						log.Printf("⏱️ %s: start=%s, time=%s",
							metric.Name,
							metric.Histogram.DataPoints[i].StartTimeUnixNano,
							metric.Histogram.DataPoints[i].TimeUnixNano,
						)
						updateHistogramDatapointTimestamps(&metric.Histogram.DataPoints[i], currentTime, defaultDiff)
					}
				}
			}
		}
	}
}

func updateGenericDatapointTimestamps(dp *common.DataPoint, now int64, fallbackDiff int64) {
	updateStringTimestamps(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now, fallbackDiff)
}

func updateHistogramDatapointTimestamps(dp *common.HistogramDataPoint, now int64, fallbackDiff int64) {
	updateStringTimestamps(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now, fallbackDiff)
}

func updateStringTimestamps(startStr *string, endStr *string, now int64, fallbackDiff int64) {
	const maxAllowedDiff = int64(10 * time.Second) // 10 seconds in nanoseconds

	var (
		startTime int64
		endTime   int64
		err1      error
		err2      error
	)

	if *startStr != "" {
		startTime, err1 = strconv.ParseInt(*startStr, 10, 64)
	} else {
		err1 = fmt.Errorf("StartTimeUnixNano is empty")
	}
	if *endStr != "" {
		endTime, err2 = strconv.ParseInt(*endStr, 10, 64)
	} else {
		err2 = fmt.Errorf("TimeUnixNano is empty")
	}

	timeDiff := fallbackDiff
	if err1 == nil && err2 == nil {
		diff := endTime - startTime
		if diff > 0 && diff < maxAllowedDiff {
			timeDiff = diff
		} else {
			log.Printf("⚠️ Clamping invalid time diff (%d ns) to default %d ns", diff, fallbackDiff)
		}
	} else {
		log.Printf("⚠️ Using fallback diff due to parse errors: %v / %v", err1, err2)
	}

	*startStr = fmt.Sprintf("%d", now)
	*endStr = fmt.Sprintf("%d", now+timeDiff)
}
