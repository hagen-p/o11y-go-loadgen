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
						UpdateInstantaneousDatapointTimestamps(&metric.Gauge.DataPoints[i], now, defaultDiff)
					}
				}
				if metric.Sum != nil {
					for i := range metric.Sum.DataPoints {
						if metric.Sum.AggregationTemporality == 2 {
							// Cumulative: preserve start time
							UpdateCumulativeDatapointTimestamp(&metric.Sum.DataPoints[i], now)
						} else {
							// Delta or unspecified: update both
							UpdateInstantaneousDatapointTimestamps(&metric.Sum.DataPoints[i], now, defaultDiff)
						}
					}
					FixSumMetric(&metric)
				}
				if metric.Histogram != nil {
					for i := range metric.Histogram.DataPoints {
						UpdateInstantaneousHistogramTimestamps(&metric.Histogram.DataPoints[i], now, defaultDiff)
					}
				}
			}
		}
	}
}

func UpdateInstantaneousDatapointTimestamps(dp *DataPoint, now int64, fallbackDiff int64) {
	UpdateStringTimestamps(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now, fallbackDiff)
}

func UpdateInstantaneousHistogramTimestamps(dp *HistogramDataPoint, now int64, fallbackDiff int64) {
	UpdateStringTimestamps(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now, fallbackDiff)
}

func UpdateCumulativeDatapointTimestamp(dp *DataPoint, now int64) {
	UpdateStringTimestampsKeepStart(&dp.StartTimeUnixNano, &dp.TimeUnixNano, now)
}

func UpdateStringTimestamps(startStr *string, endStr *string, now int64, fallbackDiff int64) {
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
	} else {
		log.Printf("⚠️ Using fallback diff due to parse errors: %v / %v", err1, err2)
	}
	*startStr = fmt.Sprintf("%d", now)
	*endStr = fmt.Sprintf("%d", now+timeDiff)
}

func UpdateStringTimestampsKeepStart(startStr *string, endStr *string, now int64) {
	*endStr = fmt.Sprintf("%d", now)
}

func FixSumMetric(metric *Metric) {
	if metric.Sum == nil {
		return
	}

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
