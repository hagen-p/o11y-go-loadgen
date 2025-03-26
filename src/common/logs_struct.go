package common

type LogFile struct {
	ResourceLogs []ResourceLog `json:"resourceLogs"`
}

type ResourceLog struct {
	Resource  Resource   `json:"resource"`
	ScopeLogs []ScopeLog `json:"scopeLogs"`
	SchemaUrl string     `json:"schemaUrl,omitempty"`
}

type ScopeLog struct {
	Scope InstrumentationScope `json:"scope"`
	Logs  []LogRecord          `json:"logRecords"`
}

type LogRecord struct {
	TimeUnixNano         string      `json:"timeUnixNano"`
	ObservedTimeUnixNano string      `json:"observedTimeUnixNano,omitempty"`
	SeverityNumber       int         `json:"severityNumber,omitempty"`
	SeverityText         string      `json:"severityText,omitempty"`
	Body                 LogBody     `json:"body"`
	Attributes           []Attribute `json:"attributes,omitempty"`
}

type LogBody struct {
	StringValue string `json:"stringValue,omitempty"`
	// You can expand to support intValue, boolValue, etc.
}
