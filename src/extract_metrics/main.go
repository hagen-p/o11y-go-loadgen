package main

import (
	"log"
        "../common"
)

func main() {
	loadConfig("config.yaml")
	processMetricsFile()
	log.Println("🏁 Processing complete.")
}
