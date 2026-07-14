package client

import (
	"fmt"
	"io"
)

const maxBufferedResponseBytes int64 = 64 << 20

func readAllLimited(r io.Reader) ([]byte, error) {
	limited := &io.LimitedReader{R: r, N: maxBufferedResponseBytes + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBufferedResponseBytes {
		return nil, fmt.Errorf("response exceeds %d MiB limit; narrow the query or use a streaming command", maxBufferedResponseBytes>>20)
	}
	return data, nil
}
