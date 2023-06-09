package cli

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func expectFileContent(t *testing.T, file, got string) {
	data, err := os.ReadFile(file)
	// Ignore windows line endings
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))

	r := require.New(t)
	r.NoError(err)
	r.Equal(got, string(data))
}

func buildTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)

	return ctx
}
