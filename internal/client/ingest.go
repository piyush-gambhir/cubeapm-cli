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
		path = "/api/v1/import/prometheus"
		contentType = "text/plain"
	case "otlp":
		path = "/v1/metrics"
		contentType = "application/x-protobuf"
	default:
		return fmt.Errorf("unsupported metrics format %q: use 'prometheus' or 'otlp'", format)
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
		path = "/v1/logs"
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
