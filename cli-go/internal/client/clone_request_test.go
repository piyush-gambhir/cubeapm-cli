package client

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestCloneRequestResendsAConsumedBody(t *testing.T) {
	req, err := http.NewRequest("POST", "https://cube.example/api/metrics/api/v1/query", bytes.NewReader([]byte("query=1")))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadAll(req.Body); err != nil {
		t.Fatal(err)
	}
	clone, err := cloneRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(clone.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "query=1" || clone.ContentLength != 7 {
		t.Fatalf("retry body %q, length %d; want %q, 7", got, clone.ContentLength, "query=1")
	}
}
