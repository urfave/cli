package cli_test

import (
	"os/exec"
	"strings"
	"testing"
)

// The core must stay small. These packages either cost hundreds of kilobytes
// or, like text/template, stop the linker from removing unused methods
// anywhere in the program. They belong in optional packages.
var banned = []string{"text/template", "html/template", "encoding/json", "net", "net/http", "golang.org/x/"}

func TestCoreDependencies(t *testing.T) {
	for _, pkg := range []string{".", "./examples/hello", "./examples/hello-i18n"} {
		t.Run(pkg, func(t *testing.T) {
			out, err := exec.Command("go", "list", "-deps", pkg).Output()
			if err != nil {
				t.Fatalf("go list: %v", err)
			}
			for _, dep := range strings.Fields(string(out)) {
				for _, b := range banned {
					if dep == b || (strings.HasSuffix(b, "/") && strings.HasPrefix(dep, b)) {
						t.Errorf("%s depends on %s", pkg, dep)
					}
				}
			}
		})
	}
}
