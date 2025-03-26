package common

// MetricsFile represents the full OTLP JSON structure.
type MetricsFile struct {
	ResourceMetrics []ResourceMetric `json:"resourceMetrics"`
}

type ResourceMetric struct {
	Resource     Resource      `json:"resource"`
	ScopeMetrics []ScopeMetric `json:"scopeMetrics"`
	SchemaUrl    string        `json:"schemaUrl,omitempty"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string    `json:"key"`
	Value AttrValue `json:"value"`
}

type AttrValue struct {
	StringValue string `json:"stringValue,omitempty"`
	// Add more types if needed (e.g., IntValue, BoolValue, etc.)
}

type ScopeMetric struct {
	Scope     InstrumentationScope `json:"scope"`
	Metrics   []Metric             `json:"metrics"`
	SchemaURL string               `json:"schemaUrl"`
}

type InstrumentationScope struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Metric struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Unit        string     `json:"unit,omitempty"`
	Sum         *Sum       `json:"sum,omitempty"`
	Gauge       *Gauge     `json:"gauge,omitempty"`
	Histogram   *Histogram `json:"histogram,omitempty"`
}

type Sum struct {
	AggregationTemporality int         `json:"aggregationTemporality"`
	IsMonotonic            bool        `json:"isMonotonic,omitempty"`
	DataPoints             []DataPoint `json:"dataPoints"`
}

type Gauge struct {
	DataPoints []DataPoint `json:"dataPoints"`
}

type Histogram struct {
	DataPoints []HistogramDataPoint `json:"dataPoints"`
}

type DataPoint struct {
	Attributes        []Attribute `json:"attributes,omitempty"`
	StartTimeUnixNano string      `json:"startTimeUnixNano"`
	TimeUnixNano      string      `json:"timeUnixNano"`
	AsInt             string      `json:"asInt,omitempty"`
	AsDouble          *float64    `json:"asDouble,omitempty"`
}

type HistogramDataPoint struct {
	StartTimeUnixNano string    `json:"startTimeUnixNano"`
	TimeUnixNano      string    `json:"timeUnixNano"`
	Count             uint64    `json:"count"`
	Sum               float64   `json:"sum"`
	BucketCounts      []uint64  `json:"bucketCounts"`
	ExplicitBounds    []float64 `json:"explicitBounds"`
}
