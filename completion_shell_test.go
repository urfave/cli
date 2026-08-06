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
	// bashWords is what bash puts in COMP_WORDS for line, with the cursor at its end.
	// Bash splits on COMP_WORDBREAKS and keeps the quoting as typed, and a driver that
	// works the words out for itself tests its own idea of that rather than the
	// script's: these were measured in bash 5.3 with a completion function that dumps
	// COMP_WORDS.
	bashWords []string
	// bashSkip says why bash is left out of this line, and is empty when it is not.
	// Leaving bashWords out is not the signal: a case that simply forgot them would
	// then quietly cover one shell fewer, which is what CLI_SHELL_TESTS_REQUIRED
	// prevents a level up, for a shell rather than for a case.
	bashSkip string
	// pwshLine is line as it would be typed in PowerShell, where it differs. A quoted
	// command name is only a command there when the call operator says so: "'app' su"
	// is a string followed by a word, not a command being completed.
	pwshLine string
	want     []string
}

func completionCases() []completionCase {
	return []completionCase{
		{
			name:      "a word being typed",
			line:      "app su",
			bashWords: []string{"app", "su"},
			want:      []string{"__complete", "su"},
		},
		{
			name:      "a fresh word",
			line:      "app sub ",
			bashWords: []string{"app", "sub", ""},
			want:      []string{"__complete", "sub", ""},
		},
		{
			name:      "a flag being typed",
			line:      "app sub --fl",
			bashWords: []string{"app", "sub", "--fl"},
			want:      []string{"__complete", "sub", "--fl"},
		},
		{
			// COMP_WORDBREAKS holds "=", so bash splits this into three words and has
			// to put them back together before asking.
			name:      "a flag holding its value",
			line:      "app --opt=va",
			bashWords: []string{"app", "--opt", "=", "va"},
			want:      []string{"__complete", "--opt=va"},
		},
		{
			// One level of quoting comes off, the way the shell takes it off before
			// handing a word to a command.
			name:      "a quoted word",
			line:      `app sub "hello world" `,
			bashWords: []string{"app", "sub", `"hello world"`, ""},
			want:      []string{"__complete", "sub", "hello world", ""},
		},
		{
			// The quote is still open, so there is no closing quote to take off with
			// it. The word is what is being typed, not one starting with a quote.
			name:      "a word whose quote is still open",
			line:      `app sub "hello wo`,
			bashWords: []string{"app", "sub", `"hello wo`},
			want:      []string{"__complete", "sub", "hello wo"},
		},
		{
			// Everything after "--" is a positional argument of whatever the command
			// runs, and the command is asked about it rather than running it.
			name:      "past a double dash",
			line:      "app exec -- git push ",
			bashWords: []string{"app", "exec", "--", "git", "push", ""},
			want:      []string{"__complete", "exec", "--", "git", "push", ""},
		},
		{
			// The command word is a word like any other: one level of quoting comes
			// off it too, or the request is sent to a command whose name holds the
			// quotes.
			name:     "a quoted command word",
			line:     "'app' su",
			pwshLine: "& 'app' su",
			// Measured the way the words above were, with a completion function that
			// records having been called: 'app' su and "app" su never reach it.
			bashSkip: "a command word holding a quote matches no compspec, so bash completes it as a file name",
			want:     []string{"__complete", "su"},
		},
		{
			// Answering a completion must not evaluate the command line. The old
			// scripts re-parsed it, so a command substitution ran on the tab key.
			name:      "a command substitution",
			line:      "app sub $(touch NOPE) ",
			bashWords: []string{"app", "sub", "$(touch NOPE)", ""},
			want:      []string{"__complete", "sub", "$(touch NOPE)", ""},
		},
	}
}

// pwshCommandLine is the command line to complete in PowerShell.
func (tc completionCase) pwshCommandLine() string {
	if tc.pwshLine != "" {
		return tc.pwshLine
	}
	return tc.line
}

// TestCompletionScriptsRequest runs the generated scripts in the shells they are
// written for and checks the request each one builds.
func TestCompletionScriptsRequest(t *testing.T) {
	// testing.Short is not read here: tests in this package add flags of their own to
	// the standard flag set, which leaves testing.Short panicking on a flag set that
	// has not been parsed. The shells run in parallel and each one skips when it is
	// not installed, so the cost of leaving it in is a few seconds on a machine that
	// has all four.
	t.Parallel()

	checkRequiredShells(t)

	for _, shell := range []string{"bash", "zsh", "fish", "pwsh"} {
		t.Run(shell, func(t *testing.T) {
			t.Parallel()

			driver := shellDrivers[shell]
			interpreter, err := exec.LookPath(driver.interpreter)
			if err != nil {
				skipMissingShell(t, shell, driver.interpreter+" is not installed")
			}

			render := shellCompletions[shell]
			require.NotNil(t, render)
			script, err := render(&Command{Name: "app", EnableShellCompletion: true}, "app")
			require.NoError(t, err)

			for _, tc := range completionCases() {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					if shell == "bash" && tc.bashSkip != "" {
						t.Skip(tc.bashSkip)
					}

					dir := t.TempDir()
					scriptPath := filepath.Join(dir, "completion."+shell)
					require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o644))

					argvPath := filepath.Join(dir, "argv")
					writeCompletionTestApp(t, dir)

					prelude := driver.prelude(t, interpreter)
					program := driver.program(scriptPath, tc)

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

// checkRequiredShells fails when CLI_SHELL_TESTS_REQUIRED names a shell this file does
// not know. The variable is there so that coverage cannot go away quietly, which a
// typo in it would undo: a name matching nothing requires nothing.
func checkRequiredShells(t *testing.T) {
	t.Helper()
	for _, name := range strings.Split(os.Getenv("CLI_SHELL_TESTS_REQUIRED"), ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := shellDrivers[name]; !ok {
			t.Fatalf("CLI_SHELL_TESTS_REQUIRED names %q, which is not a shell driven here", name)
		}
	}
}

// skipMissingShell skips a shell that is not installed, unless it is one the
// environment names as required. A skip is silent, and a machine that has none of the
// four reports the same green as one where every request is right, so a run that is
// meant to cover a shell says which ones and fails when it cannot.
//
// The shells are named rather than required as a group, so that a job requiring what
// it installs is not broken by a shell disappearing from a runner image.
func skipMissingShell(t *testing.T, shell, reason string) {
	t.Helper()
	for _, required := range strings.Split(os.Getenv("CLI_SHELL_TESTS_REQUIRED"), ",") {
		if strings.TrimSpace(required) == shell {
			t.Fatalf("%s is required by CLI_SHELL_TESTS_REQUIRED: %s", shell, reason)
		}
	}
	t.Skip(reason)
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
	program     func(scriptPath string, tc completionCase) string
}

var shellDrivers = map[string]shellDriver{
	"bash": {
		interpreter: "bash",
		args:        func(p string) []string { return []string{"-c", p} },
		// The script calls the word-splitting helpers of bash-completion, so without
		// it there is nothing to drive.
		prelude: func(t *testing.T, interpreter string) string {
			t.Helper()
			var unusable []string
			for _, p := range []string{
				"/usr/share/bash-completion/bash_completion",
				"/etc/bash_completion",
				"/opt/homebrew/share/bash-completion/bash_completion",
				"/usr/local/share/bash-completion/bash_completion",
			} {
				if _, err := os.Stat(p); err != nil {
					continue
				}
				// Finding the file is not the same as being able to use it:
				// bash-completion 2.12 and later need bash 4.2, so sourcing it in the
				// bash macOS ships leaves the helpers undefined and the script with
				// nothing to call. Ask this bash what it ends up with rather than
				// assuming that the file is enough, and go on looking when the answer
				// is no: an older one further down the list may still work.
				usable := exec.Command(interpreter, "-c", ". "+shQuote(p)+
					" >/dev/null 2>&1; declare -F _comp_initialize >/dev/null 2>&1 || declare -F _get_comp_words_by_ref >/dev/null 2>&1")
				if err := usable.Run(); err != nil {
					unusable = append(unusable, p)
					continue
				}
				return ". " + p + "\n"
			}
			if len(unusable) > 0 {
				skipMissingShell(t, "bash", interpreter+" cannot use the bash-completion in "+strings.Join(unusable, ", "))
			}
			skipMissingShell(t, "bash", "bash-completion is not installed")
			return ""
		},
		program: func(scriptPath string, tc completionCase) string {
			// COMP_WORDS and COMP_CWORD are what bash sets before it calls the
			// completion function, which is what the script reads. They are written
			// out as measured rather than worked out here: quoting them apart or
			// letting eval build them would test this driver's idea of what bash does
			// with a command line instead of the script's handling of what bash
			// actually produces, and eval would run a command substitution on the way.
			words := make([]string, 0, len(tc.bashWords))
			for _, w := range tc.bashWords {
				words = append(words, shQuote(w))
			}
			return fmt.Sprintf(`
. %s
COMP_WORDS=(%s)
COMP_CWORD=%d
COMP_LINE=%s
COMP_POINT=%d
__app_bash_autocomplete
`, shQuote(scriptPath), strings.Join(words, " "), len(tc.bashWords)-1, shQuote(tc.line), len(tc.line))
		},
	},
	"zsh": {
		interpreter: "zsh",
		args:        func(p string) []string { return []string{"-f", "-c", p} },
		prelude:     func(*testing.T, string) string { return "" },
		// The completion system is not started, since driving it needs a pseudo
		// terminal, so the parts of it the script uses stand in for it and words and
		// CURRENT are filled with zsh's own tokenizer. That last part is an assumption
		// rather than something checked: unlike the bash words, which are written out
		// as measured, these are what (z) makes of the line, which is close to what
		// the completion system would pass but not known to be identical.
		program: func(scriptPath string, tc completionCase) string {
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
`, shQuote(scriptPath), shQuote(tc.line))
		},
	},
	"fish": {
		interpreter: "fish",
		args:        func(p string) []string { return []string{"-c", p} },
		prelude:     func(*testing.T, string) string { return "" },
		// complete -C asks for the completions of a command line, which is the entry
		// point fish itself uses.
		program: func(scriptPath string, tc completionCase) string {
			return fmt.Sprintf("source %s\ncomplete -C %s\n", fishQuote(scriptPath), fishQuote(tc.line))
		},
	},
	"pwsh": {
		interpreter: "pwsh",
		args:        func(p string) []string { return []string{"-NoProfile", "-Command", p} },
		prelude:     func(*testing.T, string) string { return "" },
		// The script registers its completer under the name of the file it is in, so
		// its body is registered directly here. TabExpansion2 is what PowerShell calls
		// on the tab key.
		program: func(scriptPath string, tc completionCase) string {
			return fmt.Sprintf(`
$script = Get-Content %s -Raw
$start = $script.IndexOf('-ScriptBlock {') + '-ScriptBlock {'.Length
$body = $script.Substring($start, $script.LastIndexOf('}') - $start)
Register-ArgumentCompleter -Native -CommandName app -ScriptBlock ([scriptblock]::Create($body))
$line = %s
$null = TabExpansion2 -inputScript $line -cursorColumn $line.Length
`, pwshQuote(scriptPath), pwshQuote(tc.pwshCommandLine()))
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

// TestCompletionBashScriptTildeCommand checks that a command typed as "~/bin/app" is
// run as the path it stands for. eval expanded it as a side effect of re-parsing the
// command line, and dropping eval dropped that with it, leaving the completion looking
// for a command whose name starts with a tilde and finding nothing.
func TestCompletionBashScriptTildeCommand(t *testing.T) {
	t.Parallel()

	driver := shellDrivers["bash"]
	interpreter, err := exec.LookPath(driver.interpreter)
	if err != nil {
		skipMissingShell(t, "bash", driver.interpreter+" is not installed")
	}

	render := shellCompletions["bash"]
	require.NotNil(t, render)
	script, err := render(&Command{Name: "app", EnableShellCompletion: true}, "app")
	require.NoError(t, err)

	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, "bin"), 0o755))
	writeCompletionTestApp(t, filepath.Join(home, "bin"))

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "completion.bash")
	require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o644))
	argvPath := filepath.Join(dir, "argv")

	// The word is written with a quoted tilde so that this driver does not expand it:
	// what the script receives has to be the tilde bash puts in COMP_WORDS.
	program := fmt.Sprintf(`
. %s
COMP_WORDS=('~/bin/app' 'su')
COMP_CWORD=1
COMP_LINE='~/bin/app su'
COMP_POINT=12
__app_bash_autocomplete
`, shQuote(scriptPath))

	cmd := exec.Command(interpreter, driver.args(driver.prelude(t, interpreter)+program)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+home, "ARGV_LOG="+argvPath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "driving bash: %s", out)

	got, err := os.ReadFile(argvPath)
	require.NoError(t, err, "the completion did not run the command: %s", out)
	assert.Equal(t, []string{"__complete", "su"}, strings.Split(strings.TrimSuffix(string(got), "\n"), "\n"))
}

// TestCompletionScriptsSyntax checks that the generated scripts parse, for an app name
// holding characters a shell reads specially. A name is free to hold a "-" or a "."
// — docker-compose, golangci-lint — which a function name may hold and a variable name
// may not, so a script putting the name in a variable breaks for those apps only, and
// breaks the whole file: sourcing stops at the syntax error, before the completion is
// registered at all.
func TestCompletionScriptsSyntax(t *testing.T) {
	t.Parallel()

	// The syntax check for each shell, given the file to read.
	checks := map[string]func(string) []string{
		"bash": func(p string) []string { return []string{"-n", p} },
		"zsh":  func(p string) []string { return []string{"-n", p} },
		"fish": func(p string) []string { return []string{"-n", p} },
		"pwsh": func(p string) []string {
			return []string{
				"-NoProfile", "-Command",
				"$errors = $null; $null = [System.Management.Automation.Language.Parser]::ParseFile(" +
					pwshQuote(p) + ", [ref]$null, [ref]$errors); if ($errors) { $errors; exit 1 }",
			}
		},
	}

	// A space is left out: the function names have held the app name since long before
	// this file, and a name holding a space breaks those in bash and zsh whatever the
	// variables do.
	for _, name := range []string{"app", "my-app", "my.app"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, shell := range []string{"bash", "zsh", "fish", "pwsh"} {
				t.Run(shell, func(t *testing.T) {
					t.Parallel()

					interpreter, err := exec.LookPath(shellDrivers[shell].interpreter)
					if err != nil {
						skipMissingShell(t, shell, shellDrivers[shell].interpreter+" is not installed")
					}

					render := shellCompletions[shell]
					require.NotNil(t, render)
					script, err := render(&Command{Name: name, EnableShellCompletion: true}, name)
					require.NoError(t, err)

					dir := t.TempDir()
					p := filepath.Join(dir, "completion."+shell)
					require.NoError(t, os.WriteFile(p, []byte(script), 0o644))

					out, err := exec.Command(interpreter, checks[shell](p)...).CombinedOutput()
					require.NoError(t, err, "the %s script for %q does not parse: %s", shell, name, out)
				})
			}
		})
	}
}
