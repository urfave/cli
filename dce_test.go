package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reflectMethodSymbols builds examples/example-cli with the given tags,
// and returns the number of reachable symbols using reflect.Method or
// reflect.MethodByName with a non-constant argument. If there are any,
// the linker's dead code elimination of exported methods is disabled.
func reflectMethodSymbols(t *testing.T, tags string) int {
	t.Helper()
	out := filepath.Join(t.TempDir(), "example-cli")
	cmd := exec.Command("go", "build", "-tags", tags, "-ldflags=-dumpdep", "-o", out, "./examples/example-cli")
	deps, err := cmd.CombinedOutput()
	require.NoError(t, err, string(deps))
	return strings.Count(string(deps), "<ReflectMethod>")
}

// Using text/template (or anything else calling reflect.Method or
// reflect.MethodByName with a non-constant argument) disables linker's
// dead code elimination of exported methods. Make sure this does not
// happen with urfave_cli_no_template tag.
func TestNoTemplateTagDCE(t *testing.T) {
	// Sanity check: without the tag, DCE is disabled by text/template.
	assert.NotZero(t, reflectMethodSymbols(t, ""))

	assert.Zero(t, reflectMethodSymbols(t, "urfave_cli_no_template"))
}
