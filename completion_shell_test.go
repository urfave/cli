//go:build !windows

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file checks what the generated completion scripts actually send, by running
// them in the shells they are written for and recording the arguments the command
// receives. Asserting on the text of a script only says that it was generated; a word
// the shell splits, quotes or normalizes differently is a difference the text cannot
// show.
//
// Every shell is skipped when it is not installed, so this adds nothing to a machine
// that has none of them.

// completionCase is a command line, with the cursor at its end unless the line ends
// in a space, and the arguments the command is expected to receive for it.
type completionCase struct {
	name string
	line string
	want []string
}

func completionCases() []completionCase {
	return []completionCase{
		{
			name: "a word being typed",
			line: "app su",
			want: []string{"__complete", "su"},
		},
		{
			name: "a fresh word",
			line: "app sub ",
			want: []string{"__complete", "sub", ""},
		},
		{
			name: "a flag being typed",
			line: "app sub --fl",
			want: []string{"__complete", "sub", "--fl"},
		},
		{
			// COMP_WORDBREAKS holds "=", so bash splits this into three words and has
			// to put them back together before asking.
			name: "a flag holding its value",
			line: "app --opt=va",
			want: []string{"__complete", "--opt=va"},
		},
		{
			// One level of quoting comes off, the way the shell takes it off before
			// handing a word to a command.
			name: "a quoted word",
			line: `app sub "hello world" `,
			want: []string{"__complete", "sub", "hello world", ""},
		},
		{
			// The quote is still open, so there is no closing quote to take off with
			// it. The word is what is being typed, not one starting with a quote.
			name: "a word whose quote is still open",
			line: `app sub "hello wo`,
			want: []string{"__complete", "sub", "hello wo"},
		},
		{
			// Everything after "--" is a positional argument of whatever the command
			// runs, and the command is asked about it rather than running it.
			name: "past a double dash",
			line: "app exec -- git push ",
			want: []string{"__complete", "exec", "--", "git", "push", ""},
		},
		{
			// Answering a completion must not evaluate the command line. The old
			// scripts re-parsed it, so a command substitution ran on the tab key.
			name: "a command substitution",
			line: "app sub $(touch NOPE) ",
			want: []string{"__complete", "sub", "$(touch NOPE)", ""},
		},
	}
}

// TestCompletionScriptsRequest runs the generated scripts in the shells they are
// written for and checks the request each one builds.
func TestCompletionScriptsRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("driving four shells takes seconds, not milliseconds")
	}

	t.Parallel()

	for _, shell := range []string{"bash", "zsh", "fish", "pwsh"} {
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			driver := shellDrivers[shell]
			interpreter, err := exec.LookPath(driver.interpreter)
			if err != nil {
				t.Skipf("%s is not installed", driver.interpreter)
			}

			render := shellCompletions[shell]
			require.NotNil(t, render)
			script, err := render(&Command{Name: "app", EnableShellCompletion: true}, "app")
			require.NoError(t, err)

			for _, tc := range completionCases() {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					dir := t.TempDir()
					scriptPath := filepath.Join(dir, "completion."+shell)
					require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o644))

					argvPath := filepath.Join(dir, "argv")
					writeCompletionTestApp(t, dir)

					prelude := driver.prelude(t, interpreter)
					program := driver.program(scriptPath, tc.line)

					cmd := exec.Command(interpreter, driver.args(prelude+program)...)
					cmd.Dir = dir
					cmd.Env = append(os.Environ(),
						"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
						"ARGV_LOG="+argvPath,
					)
					out, err := cmd.CombinedOutput()
					require.NoError(t, err, "driving %s: %s", shell, out)

					got, err := os.ReadFile(argvPath)
					require.NoError(t, err, "the completion did not run the command: %s", out)
					assert.Equal(t, tc.want, strings.Split(strings.TrimSuffix(string(got), "\n"), "\n"))

					assert.NoFileExists(t, filepath.Join(dir, "NOPE"),
						"the command line must not be evaluated to answer a completion")
				})
			}
		})
	}
}

// writeCompletionTestApp writes the command the completion scripts ask, which records
// the arguments it receives and answers with one candidate.
func writeCompletionTestApp(t *testing.T, dir string) {
	t.Helper()
	app := "#!/bin/sh\n: > \"$ARGV_LOG\"\nfor a in \"$@\"; do printf '%s\\n' \"$a\" >> \"$ARGV_LOG\"; done\necho candidate\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app"), []byte(app), 0o755))
}

// shellDriver runs a completion the way its shell would, without a terminal. Each
// shell offers its own way in: what they have in common is that the script under test
// is sourced and the completion for a command line is asked for.
type shellDriver struct {
	interpreter string
	args        func(program string) []string
	prelude     func(t *testing.T, interpreter string) string
	program     func(scriptPath, line string) string
}

var shellDrivers = map[string]shellDriver{
	"bash": {
		interpreter: "bash",
		args:        func(p string) []string { return []string{"-c", p} },
		// The script calls the word-splitting helpers of bash-completion, so without
		// it there is nothing to drive.
		prelude: func(t *testing.T, _ string) string {
			t.Helper()
			for _, p := range []string{
				"/usr/share/bash-completion/bash_completion",
				"/etc/bash_completion",
				"/opt/homebrew/share/bash-completion/bash_completion",
				"/usr/local/share/bash-completion/bash_completion",
			} {
				if _, err := os.Stat(p); err == nil {
					return ". " + p + "\n"
				}
			}
			t.Skip("bash-completion is not installed")
			return ""
		},
		program: func(scriptPath, line string) string {
			// COMP_WORDS and COMP_CWORD are what bash sets before it calls the
			// completion function, which is what the script reads.
			return fmt.Sprintf(`
. %s
line=%s
eval "COMP_WORDS=($line)"
[ "${line: -1}" = " " ] && COMP_WORDS+=("")
COMP_CWORD=$(( ${#COMP_WORDS[@]} - 1 ))
COMP_LINE="$line"
COMP_POINT=${#line}
__app_bash_autocomplete
`, shQuote(scriptPath), shQuote(line))
		},
	},
	"zsh": {
		interpreter: "zsh",
		args:        func(p string) []string { return []string{"-f", "-c", p} },
		prelude:     func(*testing.T, string) string { return "" },
		// The completion system is not started, so the parts of it the script uses
		// stand in for it: what is under test is the request the script builds from
		// words and CURRENT, which zsh fills the same way here.
		program: func(scriptPath, line string) string {
			return fmt.Sprintf(`
compdef() { : }
_describe() { : }
_files() { : }
. %s
line=%s
words=("${(z)line}")
[[ "$line" == *" " ]] && words+=("")
CURRENT=$#words
_app
`, shQuote(scriptPath), shQuote(line))
		},
	},
	"fish": {
		interpreter: "fish",
		args:        func(p string) []string { return []string{"-c", p} },
		prelude:     func(*testing.T, string) string { return "" },
		// complete -C asks for the completions of a command line, which is the entry
		// point fish itself uses.
		program: func(scriptPath, line string) string {
			return fmt.Sprintf("source %s\ncomplete -C %s\n", fishQuote(scriptPath), fishQuote(line))
		},
	},
	"pwsh": {
		interpreter: "pwsh",
		args:        func(p string) []string { return []string{"-NoProfile", "-Command", p} },
		prelude:     func(*testing.T, string) string { return "" },
		// The script registers its completer under the name of the file it is in, so
		// its body is registered directly here. TabExpansion2 is what PowerShell calls
		// on the tab key.
		program: func(scriptPath, line string) string {
			return fmt.Sprintf(`
$script = Get-Content %s -Raw
$start = $script.IndexOf('-ScriptBlock {') + '-ScriptBlock {'.Length
$body = $script.Substring($start, $script.LastIndexOf('}') - $start)
Register-ArgumentCompleter -Native -CommandName app -ScriptBlock ([scriptblock]::Create($body))
$line = %s
$null = TabExpansion2 -inputScript $line -cursorColumn $line.Length
`, pwshQuote(scriptPath), pwshQuote(line))
		},
	},
}

// shQuote quotes s for bash and zsh.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// fishQuote quotes s for fish.
func fishQuote(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(s) + "'"
}

// pwshQuote quotes s for PowerShell.
func pwshQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
