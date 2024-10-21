package cli

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestTracing(t *testing.T) {
	olderr := os.Stderr
	defer func() { os.Stderr = olderr }()

	file, err := os.CreateTemp(os.TempDir(), "cli*")
	assert.NoError(t, err)
	os.Stderr = file

	// Note we cant really set the env since the isTracingOn
	// is read at module startup so any changes mid code
	// wont take effect
	isTracingOn = false
	tracef("something")

	isTracingOn = true
	tracef("foothing")

	assert.NoError(t, file.Close())

	b, err := os.ReadFile(file.Name())
	assert.NoError(t, err)

	assert.Contains(t, string(b), "foothing")
	assert.NotContains(t, string(b), "something")
}
