package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func expectFileContent(t *testing.T, file, got string) {
	data, err := os.ReadFile(file)
	// Ignore windows line endings
	// TODO: Replace with bytes.ReplaceAll when support for Go 1.11 is dropped
	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)

	r := require.New(t)
	r.NoError(err)
	r.Equal(got, string(data))
}
