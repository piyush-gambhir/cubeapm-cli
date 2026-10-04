package update

import (
	"fmt"
	"io"
)

const maxReleaseArtifactBytes int64 = 256 << 20

// copyLimited copies src to dst and fails when src holds more than limit bytes.
func copyLimited(dst io.Writer, src io.Reader, limit int64) error {
	written, err := io.CopyN(dst, src, limit+1)
	if err != nil && err != io.EOF {
		return err
	}
	if written > limit {
		return fmt.Errorf("download exceeds %d MiB limit", limit>>20)
	}
	return nil
}
