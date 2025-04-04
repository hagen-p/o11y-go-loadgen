package common

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

func UpdateTimestamps(metricsFile *MetricsFile) {
	now := time.Now().UnixNano()
	const defaultDiff = int64(5)
	for _, rm := range metricsFile.ResourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			for _, metric := range sm.Metrics {
				if metric.Gauge != nil {
					for i := range metric.Gauge.DataPoints {
						UpdateGenericDatapointTimestamps(&metric.Gauge.DataPoints[i], now, defaultDiff)
					}
				}
				if metric.Sum != nil {
					for i := range metric.Sum.DataPoints {
						UpdateGenericDatapointTimestamps(&metric.Sum.DataPoints[i], now, defaultDiff)
					}
					FixSumMetric(&metric)
				}
				if metric.Histogram != nil {
					for i := range metric.Histogram.DataPoints {
						UpdateHistogramDatapointTimestamps(&metric.Histogram.DataPoints[i], now, defaultDiff)
					}
				}
			}
		}
	}
}

func UpdateGenericDatapointTimestamps(dp *DataPoint, now int64, fallbackDiff int64) {
	UpdateStringTimestamps(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now, fallbackDiff)
}

func UpdateHistogramDatapointTimestamps(dp *HistogramDataPoint, now int64, fallbackDiff int64) {
	UpdateStringTimestamps(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now, fallbackDiff)
}

func UpdateStringTimestamps(startStr *string, endStr *string, now int64, fallbackDiff int64) {
	// const maxAllowedDiff = int64(10 * time.Second) // ← No longer used
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
		if diff > 0 {
			timeDiff = diff
		} else {
			log.Printf("⚠️ Clamping non-positive time diff (%d ns) to default %d ns", diff, fallbackDiff)
		}

		// Old logic that clamped timeDiff if it was too large:
		// if diff > 0 && diff < maxAllowedDiff {
		// 	timeDiff = diff
		// } else {
		// 	log.Printf("⚠️ Clamping invalid time diff (%d ns) to default %d ns", diff, fallbackDiff)
		// }
	} else {
		log.Printf("⚠️ Using fallback diff due to parse errors: %v / %v", err1, err2)
	}
	*startStr = fmt.Sprintf("%d", now)
	*endStr = fmt.Sprintf("%d", now+timeDiff)
}
func FixSumMetric(metric *Metric) {
	if metric.Sum == nil {
		return
	}

	// Nothing to preserve; just leave as-is and move to timestamps
	now := time.Now().UnixNano()
	for i := range metric.Sum.DataPoints {
		dp := &metric.Sum.DataPoints[i]
		ensureReasonableTimestamps(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now)
	}
}

func ensureReasonableTimestamps(startStr *string, endStr *string, now int64) {
	start, err1 := strconv.ParseInt(*startStr, 10, 64)
	end, err2 := strconv.ParseInt(*endStr, 10, 64)

	if err1 != nil || err2 != nil || end <= start {
		log.Printf("⚠️ Fixing broken timestamps: %v / %v", err1, err2)
		start = now
		end = now + int64(5*time.Second)
	}

	*startStr = strconv.FormatInt(start, 10)
	*endStr = strconv.FormatInt(end, 10)
}
