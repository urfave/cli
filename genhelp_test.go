package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Make sure help_notemplate_gen.go is up to date with the templates.
func TestGeneratedHelpUpToDate(t *testing.T) {
	out := filepath.Join(t.TempDir(), "gen.go")
	cmd := exec.Command("go", "run", "./internal/genhelp", "-o", out)
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())

	got, err := os.ReadFile(out)
	require.NoError(t, err)
	expectFileContent(t, "help_notemplate_gen.go", string(got))
}
