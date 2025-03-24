package common

import (
	"encoding/json"
	"log"
)

// DeepCopyMetricsFile returns a full deep copy of the given MetricsFile
func DeepCopyMetricsFile(orig MetricsFile) MetricsFile {
	var copy MetricsFile

	data, err := json.Marshal(orig)
	if err != nil {
		log.Printf("❌ Failed to marshal metrics for deep copy: %v", err)
		return copy
	}

	if err := json.Unmarshal(data, &copy); err != nil {
		log.Printf("❌ Failed to unmarshal deep copy: %v", err)
	}

	return copy
}
