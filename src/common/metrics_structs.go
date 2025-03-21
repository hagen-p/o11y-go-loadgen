package common

import "encoding/json"

// MetricsFile represents a single OTLP JSON-formatted metric set
type MetricsFile struct {
	Resource    Resource    `json:"resource"`
	ScopeMetric ScopeMetric `json:"scopeMetric"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value struct {
		StringValue string `json:"stringValue"`
	} `json:"value"`
}

type ScopeMetric struct {
	Metrics   []Metric `json:"metrics"`
	SchemaURL string   `json:"schemaUrl"`
	Scope     Scope    `json:"scope"`
}

type Scope struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Metric struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit,omitempty"`
	Gauge       struct {
		DataPoints []DataPoint `json:"dataPoints"`
	} `json:"gauge,omitempty"`
}

type DataPoint struct {
	StartTimeUnixNano string          `json:"startTimeUnixNano"`
	TimeUnixNano      string          `json:"timeUnixNano"`
	AsInt             json.RawMessage `json:"asInt,omitempty"`
	AsDouble          *float64        `json:"asDouble,omitempty"`
}
