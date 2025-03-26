package common

import (
	"log"
	"strconv"

	collectorpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// Convert entire MetricsFile to OTLP ExportMetricsServiceRequest
func ToOTLPRequest(metricsFile MetricsFile) *collectorpb.ExportMetricsServiceRequest {
	var otlpResourceMetrics []*metricpb.ResourceMetrics

	for _, rm := range metricsFile.ResourceMetrics {
		var otlpScopeMetrics []*metricpb.ScopeMetrics
		for _, sm := range rm.ScopeMetrics {
			otlpScopeMetrics = append(otlpScopeMetrics, &metricpb.ScopeMetrics{
				Scope:     ToOTLPScope(sm.Scope),
				SchemaUrl: sm.SchemaURL,
				Metrics:   ToOTLPMetrics(sm.Metrics),
			})
		}

		otlpResourceMetrics = append(otlpResourceMetrics, &metricpb.ResourceMetrics{
			Resource: &resourcepb.Resource{
				Attributes: ToOTLPAttributes(rm.Resource.Attributes),
			},
			ScopeMetrics: otlpScopeMetrics,
			SchemaUrl:    rm.SchemaUrl,
		})
	}

	return &collectorpb.ExportMetricsServiceRequest{
		ResourceMetrics: otlpResourceMetrics,
	}
}

func ToOTLPAttributes(attrs []Attribute) []*commonpb.KeyValue {
	var result []*commonpb.KeyValue
	for _, attr := range attrs {
		result = append(result, &commonpb.KeyValue{
			Key: attr.Key,
			Value: &commonpb.AnyValue{
				Value: &commonpb.AnyValue_StringValue{
					StringValue: attr.Value.StringValue,
				},
			},
		})
	}
	return result
}

func ToOTLPScope(scope InstrumentationScope) *commonpb.InstrumentationScope {
	return &commonpb.InstrumentationScope{
		Name:    scope.Name,
		Version: scope.Version,
	}
}

func ToOTLPMetrics(metrics []Metric) []*metricpb.Metric {
	var result []*metricpb.Metric
	for _, m := range metrics {
		metric := &metricpb.Metric{
			Name:        m.Name,
			Description: m.Description,
			Unit:        m.Unit,
		}

		if m.Gauge != nil && len(m.Gauge.DataPoints) > 0 {
			gauge := &metricpb.Gauge{}
			for _, dp := range m.Gauge.DataPoints {
				dataPoint := toOTLPDataPoint(dp)
				if dataPoint != nil {
					gauge.DataPoints = append(gauge.DataPoints, dataPoint)
				}
			}
			metric.Data = &metricpb.Metric_Gauge{Gauge: gauge}
		} else if m.Sum != nil && len(m.Sum.DataPoints) > 0 {
			sum := &metricpb.Sum{
				AggregationTemporality: metricpb.AggregationTemporality(m.Sum.AggregationTemporality),
				IsMonotonic:            m.Sum.IsMonotonic,
			}
			for _, dp := range m.Sum.DataPoints {
				dataPoint := toOTLPDataPoint(dp)
				if dataPoint != nil {
					sum.DataPoints = append(sum.DataPoints, dataPoint)
				}
			}
			metric.Data = &metricpb.Metric_Sum{Sum: sum}
		} else if m.Histogram != nil && len(m.Histogram.DataPoints) > 0 {
			hist := &metricpb.Histogram{}
			for _, dp := range m.Histogram.DataPoints {
				var sumPtr *float64
				if dp.Sum != 0 {
					sumPtr = &dp.Sum
				}
				hdp := &metricpb.HistogramDataPoint{
					StartTimeUnixNano: parseUint(dp.StartTimeUnixNano),
					TimeUnixNano:      parseUint(dp.TimeUnixNano),
					Count:             dp.Count,
					Sum:               sumPtr,
					BucketCounts:      dp.BucketCounts,
					ExplicitBounds:    dp.ExplicitBounds,
					Attributes:        []*commonpb.KeyValue{},
				}
				hist.DataPoints = append(hist.DataPoints, hdp)
			}
			metric.Data = &metricpb.Metric_Histogram{Histogram: hist}
		}

		result = append(result, metric)
	}
	return result
}

func toOTLPDataPoint(dp DataPoint) *metricpb.NumberDataPoint {
	dataPoint := &metricpb.NumberDataPoint{
		StartTimeUnixNano: parseUint(dp.StartTimeUnixNano),
		TimeUnixNano:      parseUint(dp.TimeUnixNano),
		Attributes:        []*commonpb.KeyValue{},
	}

	if dp.AsDouble != nil {
		dataPoint.Value = &metricpb.NumberDataPoint_AsDouble{
			AsDouble: *dp.AsDouble,
		}
	} else if dp.AsInt != "" {
		if val, err := strconv.ParseInt(dp.AsInt, 10, 64); err == nil {
			dataPoint.Value = &metricpb.NumberDataPoint_AsInt{
				AsInt: val,
			}
		} else {
			log.Printf("⚠️ Failed to parse AsInt value '%s': %v", dp.AsInt, err)
		}
	}

	if dataPoint.Value == nil {
		log.Printf("⚠️ Dropping invalid datapoint: no value (asInt/asDouble)")
		return nil
	}

	return dataPoint
}

func parseUint(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}
