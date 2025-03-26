package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hagen-p/o11y-go-loadgen/src/common"
)

func main() {

	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Display usage information")

	flag.BoolVar(&common.DebugEnabled, "d", false, "Enable debug output")
	flag.BoolVar(&common.InfoEnabled, "I", false, "Enable info-level logs to stdout")

	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage: metrics_loadgen [options]")
		fmt.Println("Options:")
		fmt.Println("  --config=<path>  Specify the configuration file (default: config.yaml)")
		fmt.Println("  -d               Enable debug logs")
		fmt.Println("  -I               Enable info logs to stdout")
		fmt.Println("  -h               Display this help message")
		os.Exit(0)
	}
	common.InitLogging()
	common.LoadConfig(*configPath)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("🛑 Stopping JSON processing...")
		os.Exit(0)
	}()

	for {
		//processFiles()
		processSingleFile()
		time.Sleep(10 * time.Second)
		if common.DebugEnabled {
			os.Exit(0)
		}
	}
}
