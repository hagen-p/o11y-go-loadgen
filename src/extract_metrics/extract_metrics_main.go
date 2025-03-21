package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Display usage information")
	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage: metrics_loadgen [options]")
		fmt.Println("Options:")
		fmt.Println("  --config=<path>  Specify the configuration file (default: config.yaml)")
		fmt.Println("  -h               Display this help message")
		os.Exit(0)
	}

	loadConfig(*configPath)
	common.ProcessMetricsFile()
	log.Println("🏁 Processing complete.")
}
