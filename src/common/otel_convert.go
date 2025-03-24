package common

import (
	"strconv"

	collectorpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// Convert entire MetricsFile to OTLP ExportMetricsServiceRequest
func ToOTLPRequest(metricsFile MetricsFile) *collectorpb.ExportMetricsServiceRequest {
	return &collectorpb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricpb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: ToOTLPAttributes(metricsFile.Resource.Attributes),
				},
				ScopeMetrics: []*metricpb.ScopeMetrics{
					{
						Scope:     ToOTLPScope(metricsFile.ScopeMetric.Scope),
						SchemaUrl: metricsFile.ScopeMetric.SchemaURL,
						Metrics:   ToOTLPMetrics(metricsFile.ScopeMetric.Metrics),
					},
				},
			},
		},
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

func ToOTLPScope(scope Scope) *commonpb.InstrumentationScope {
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

		if len(m.Gauge.DataPoints) > 0 {
			gauge := &metricpb.Gauge{}
			for _, dp := range m.Gauge.DataPoints {
				dataPoint := &metricpb.NumberDataPoint{
					StartTimeUnixNano: parseUint(dp.StartTimeUnixNano),
					TimeUnixNano:      parseUint(dp.TimeUnixNano),
					Attributes:        []*commonpb.KeyValue{},
				}

				if dp.AsDouble != nil {
					dataPoint.Value = &metricpb.NumberDataPoint_AsDouble{
						AsDouble: *dp.AsDouble,
					}
				} else if dp.AsInt != nil {
					dataPoint.Value = &metricpb.NumberDataPoint_AsInt{
						AsInt: int64(*dp.AsInt),
					}
				}

				gauge.DataPoints = append(gauge.DataPoints, dataPoint)
			}
			metric.Data = &metricpb.Metric_Gauge{Gauge: gauge}
		}

		result = append(result, metric)
	}
	return result
}

func parseUint(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}
