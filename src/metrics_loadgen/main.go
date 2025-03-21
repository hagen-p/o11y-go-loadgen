package main

import (
        "../common"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	log.Printf("📂 Monitoring directory: %s", InputDir)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("🛑 Stopping JSON processing...")
		os.Exit(0)
	}()

	for {
		processFiles()
		time.Sleep(10 * time.Second)
	}
}
