package main

import (
	"log"
)

func main() {
	loadConfig("config.yaml")
	processMetricsFile()
	log.Println("🏁 Processing complete.")
}
