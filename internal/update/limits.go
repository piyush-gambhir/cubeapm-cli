package update

import (
	"fmt"
	"io"
)

const maxReleaseArtifactBytes int64 = 256 << 20

func copyUpdatePayload(dst io.Writer, src io.Reader) error {
	written, err := io.CopyN(dst, src, maxReleaseArtifactBytes+1)
	if err != nil && err != io.EOF {
		return err
	}
	if written > maxReleaseArtifactBytes {
		return fmt.Errorf("release artifact exceeds %d MiB limit", maxReleaseArtifactBytes>>20)
	}
	return nil
}
