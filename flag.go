package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const defaultPlaceholder = "value"

var (
	defaultSliceFlagSeparator       = ","
	defaultMapFlagKeyValueSeparator = "="
	disableSliceFlagSeparator       = false
)

var (
	slPfx = fmt.Sprintf("sl:::%d:::", time.Now().UTC().UnixNano())

	commaWhitespace = regexp.MustCompile("[, ]+.*")
)

// GenerateShellCompletionFlag enables shell completion
var GenerateShellCompletionFlag Flag = &BoolFlag{
	Name:   "generate-shell-completion",
	Hidden: true,
}

// VersionFlag prints the version for the application
var VersionFlag Flag = &BoolFlag{
	Name:        "version",
	Aliases:     []string{"v"},
	Usage:       "print the version",
	HideDefault: true,
	Local:       true,
}

// HelpFlag prints the help for all commands and subcommands.
// Set to nil to disable the flag.  The subcommand
// will still be added unless HideHelp or HideHelpCommand is set to true.
var HelpFlag Flag = &BoolFlag{
	Name:        "help",
	Aliases:     []string{"h"},
	Usage:       "show help",
	HideDefault: true,
	Local:       true,
}

// FlagStringer converts a flag definition to a string. This is used by help
// to display a flag.
var FlagStringer FlagStringFunc = stringifyFlag

// Serializer is used to circumvent the limitations of flag.FlagSet.Set
type Serializer interface {
	Serialize() string
}

// FlagNamePrefixer converts a full flag name and its placeholder into the help
// message flag prefix. This is used by the default FlagStringer.
var FlagNamePrefixer FlagNamePrefixFunc = prefixedNames

// FlagEnvHinter annotates flag help message with the environment variable
// details. This is used by the default FlagStringer.
var FlagEnvHinter FlagEnvHintFunc = withEnvHint

// FlagFileHinter annotates flag help message with the environment variable
// details. This is used by the default FlagStringer.
var FlagFileHinter FlagFileHintFunc = withFileHint

// FlagsByName is a slice of Flag.
type FlagsByName []Flag

func (f FlagsByName) Len() int {
	return len(f)
}

func (f FlagsByName) Less(i, j int) bool {
	if len(f[j].Names()) == 0 {
		return false
	} else if len(f[i].Names()) == 0 {
		return true
	}
	return lexicographicLess(f[i].Names()[0], f[j].Names()[0])
}

func (f FlagsByName) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// ActionableFlag is an interface that wraps Flag interface and RunAction operation.
type ActionableFlag interface {
	RunAction(context.Context, *Command) error
}

// Flag is a common interface related to parsing flags in cli.
// For more advanced flag parsing techniques, it is recommended that
// this interface be implemented.
type Flag interface {
	fmt.Stringer

	// Apply Flag settings to the given flag set
	Apply(*flag.FlagSet) error

	// All possible names for this flag
	Names() []string

	// Whether the flag has been set or not
	IsSet() bool
}

// RequiredFlag is an interface that allows us to mark flags as required
// it allows flags required flags to be backwards compatible with the Flag interface
type RequiredFlag interface {
	// whether the flag is a required flag or not
	IsRequired() bool
}

// DocGenerationFlag is an interface that allows documentation generation for the flag
type DocGenerationFlag interface {
	// TakesValue returns true if the flag takes a value, otherwise false
	TakesValue() bool

	// GetUsage returns the usage string for the flag
	GetUsage() string

	// GetValue returns the flags value as string representation and an empty
	// string if the flag takes no value at all.
	GetValue() string

	// GetDefaultText returns the default text for this flag
	GetDefaultText() string

	// GetEnvVars returns the env vars for this flag
	GetEnvVars() []string

	// IsDefaultVisible returns whether the default value should be shown in
	// help text
	IsDefaultVisible() bool
}

// DocGenerationMultiValueFlag extends DocGenerationFlag for slice/map based flags.
type DocGenerationMultiValueFlag interface {
	DocGenerationFlag

	// IsMultiValueFlag returns true for flags that can be given multiple times.
	IsMultiValueFlag() bool
}

// Countable is an interface to enable detection of flag values which support
// repetitive flags
type Countable interface {
	Count() int
}

// VisibleFlag is an interface that allows to check if a flag is visible
type VisibleFlag interface {
	// IsVisible returns true if the flag is not hidden, otherwise false
	IsVisible() bool
}

// CategorizableFlag is an interface that allows us to potentially
// use a flag in a categorized representation.
type CategorizableFlag interface {
	// Returns the category of the flag
	GetCategory() string

	// Sets the category of the flag
	SetCategory(string)
}

// LocalFlag is an interface to enable detection of flags which are local
// to current command
type LocalFlag interface {
	IsLocal() bool
}

func newFlagSet(name string, flags []Flag) (*flag.FlagSet, error) {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		if err := f.Apply(set); err != nil {
			return nil, err
		}
	}

	set.SetOutput(io.Discard)

	return set, nil
}

func visibleFlags(fl []Flag) []Flag {
	var visible []Flag
	for _, f := range fl {
		if vf, ok := f.(VisibleFlag); ok && vf.IsVisible() {
			visible = append(visible, f)
		}
	}
	return visible
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}

// Returns the placeholder, if any, and the unquoted usage string.
func unquoteUsage(usage string) (string, string) {
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name := usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break
		}
	}
	return "", usage
}

func prefixedNames(names []string, placeholder string) string {
	var prefixed string
	for i, name := range names {
		if name == "" {
			continue
		}

		prefixed += prefixFor(name) + name
		if placeholder != "" {
			prefixed += " " + placeholder
		}
		if i < len(names)-1 {
			prefixed += ", "
		}
	}
	return prefixed
}

func envFormat(envVars []string, prefix, sep, suffix string) string {
	if len(envVars) > 0 {
		return fmt.Sprintf(" [%s%s%s]", prefix, strings.Join(envVars, sep), suffix)
	}
	return ""
}

func defaultEnvFormat(envVars []string) string {
	return envFormat(envVars, "$", ", $", "")
}

func withEnvHint(envVars []string, str string) string {
	envText := ""
	if runtime.GOOS != "windows" || os.Getenv("PSHOME") != "" {
		envText = defaultEnvFormat(envVars)
	} else {
		envText = envFormat(envVars, "%", "%, %", "%")
	}
	return str + envText
}

func FlagNames(name string, aliases []string) []string {
	var ret []string

	for _, part := range append([]string{name}, aliases...) {
		// v1 -> v2 migration warning zone:
		// Strip off anything after the first found comma or space, which
		// *hopefully* makes it a tiny bit more obvious that unexpected behavior is
		// caused by using the v1 form of stringly typed "Name".
		ret = append(ret, commaWhitespace.ReplaceAllString(part, ""))
	}

	return ret
}

func withFileHint(filePath, str string) string {
	fileText := ""
	if filePath != "" {
		fileText = fmt.Sprintf(" [%s]", filePath)
	}
	return str + fileText
}

func formatDefault(format string) string {
	return " (default: " + format + ")"
}

func stringifyFlag(f Flag) string {
	// enforce DocGeneration interface on flags to avoid reflection
	df, ok := f.(DocGenerationFlag)
	if !ok {
		return ""
	}
	placeholder, usage := unquoteUsage(df.GetUsage())
	needsPlaceholder := df.TakesValue()

	if needsPlaceholder && placeholder == "" {
		placeholder = defaultPlaceholder
	}

	defaultValueString := ""

	// don't print default text for required flags
	if rf, ok := f.(RequiredFlag); !ok || !rf.IsRequired() {
		isVisible := df.IsDefaultVisible()
		if s := df.GetDefaultText(); isVisible && s != "" {
			defaultValueString = fmt.Sprintf(formatDefault("%s"), s)
		}
	}

	usageWithDefault := strings.TrimSpace(usage + defaultValueString)

	pn := prefixedNames(f.Names(), placeholder)
	sliceFlag, ok := f.(DocGenerationMultiValueFlag)
	if ok && sliceFlag.IsMultiValueFlag() {
		pn = pn + " [ " + pn + " ]"
	}

	return withEnvHint(df.GetEnvVars(), fmt.Sprintf("%s\t%s", pn, usageWithDefault))
}

func hasFlag(flags []Flag, fl Flag) bool {
	for _, existing := range flags {
		if fl == existing {
			return true
		}
	}

	return false
}

func flagSplitMultiValues(val string) []string {
	if disableSliceFlagSeparator {
		return []string{val}
	}

	return strings.Split(val, defaultSliceFlagSeparator)
}
