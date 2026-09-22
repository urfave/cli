package cli

import (
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Using text/template disables linker's dead code elimination of exported
// methods, so make sure it is not linked in with urfave_cli_no_template tag.
func TestNoTemplateTagDeps(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-tags", "urfave_cli_no_template", ".").Output()
	require.NoError(t, err)
	deps := strings.Fields(string(out))
	require.NotEmpty(t, deps)
	require.False(t, slices.Contains(deps, "text/template"), "text/template is linked in with urfave_cli_no_template tag")
}
