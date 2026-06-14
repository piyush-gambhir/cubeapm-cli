package client

import (
	"fmt"
	"io"
)

// IngestMetrics sends metrics data to the ingest port.
// Supported formats: prometheus, otlp
func (c *Client) IngestMetrics(format string, data io.Reader) error {
	var path string
	var contentType string

	switch format {
	case "prometheus":
		// Prometheus text exposition format (CubeAPM docs: POST /api/metrics/v1/save).
		path = "/api/metrics/v1/save"
		contentType = "text/plain"
	case "otlp":
		// OpenTelemetry metrics (CubeAPM docs: POST /api/metrics/v1/save/otlp).
		path = "/api/metrics/v1/save/otlp"
		contentType = "application/x-protobuf"
	case "remote-write", "remotewrite":
		// Prometheus remote write (CubeAPM docs: POST /api/metrics/api/v1/write).
		path = "/api/metrics/api/v1/write"
		contentType = "application/x-protobuf"
	default:
		return fmt.Errorf("unsupported metrics format %q: use 'prometheus', 'otlp', or 'remote-write'", format)
	}

	resp, err := c.postRaw(c.ingestBaseURL, path, contentType, data)
	if err != nil {
		return fmt.Errorf("ingesting metrics: %w", err)
	}
	defer resp.Body.Close()

	return c.checkResponse(resp)
}

// IngestLogs sends log data to the ingest port.
// Supported formats: jsonline, otlp, loki, elastic
func (c *Client) IngestLogs(format string, data io.Reader) error {
	var path string
	var contentType string

	switch format {
	case "jsonline":
		path = "/api/logs/insert/jsonline"
		contentType = "application/stream+json"
	case "otlp":
		// OpenTelemetry logs (CubeAPM docs: POST /api/logs/insert/opentelemetry/v1/logs).
		path = "/api/logs/insert/opentelemetry/v1/logs"
		contentType = "application/x-protobuf"
	case "loki":
		path = "/api/logs/insert/loki/api/v1/push"
		contentType = "application/json"
	case "elastic":
		path = "/api/logs/insert/elasticsearch/_bulk"
		contentType = "application/x-ndjson"
	default:
		return fmt.Errorf("unsupported logs format %q: use 'jsonline', 'otlp', 'loki', or 'elastic'", format)
	}

	resp, err := c.postRaw(c.ingestBaseURL, path, contentType, data)
	if err != nil {
		return fmt.Errorf("ingesting logs: %w", err)
	}
	defer resp.Body.Close()

	return c.checkResponse(resp)
}
