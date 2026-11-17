package cli

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

const defaultPlaceholder = "value"

const (
	defaultSliceFlagSeparator       = ","
	defaultMapFlagKeyValueSeparator = "="
	disableSliceFlagSeparator       = false
)

var slPfx = sync.OnceValue(func() string {
	return fmt.Sprintf("sl:::%d:::", time.Now().UTC().UnixNano())
})

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

// FlagFileHinter annotates flag help message with the file path details. It
// is not used by the default FlagStringer.
var FlagFileHinter FlagFileHintFunc = withFileHint

// FlagsByName is a slice of Flag.
type FlagsByName []Flag

func (f FlagsByName) Len() int {
	return len(f)
}

func (f FlagsByName) Less(i, j int) bool {
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

	// Retrieve the value of the Flag
	Get() any

	// Lifecycle methods.
	// flag callback prior to parsing
	PreParse() error

	// flag callback post parsing
	PostParse() error

	// Apply Flag settings to the given flag set
	Set(string, string) error

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

// DeprecatedFlag is an interface that allows us to mark flags as deprecated
type DeprecatedFlag interface {
	// GetDeprecated returns the deprecation message, or an empty string
	// if the flag is not deprecated
	GetDeprecated() string
}

// SensitiveFlag is an interface that allows us to mark flags whose values
// are secrets, such as passwords or tokens. The value of a sensitive flag
// is not shown in help output or error messages, and is returned by
// [Command.SensitiveValues] so that it can be removed from other output.
//
// Encoding a command or a flag as JSON includes the default value of a
// sensitive flag, but never the value it was set to. To keep this so for
// flag types whose value is a pointer, such as [GenericFlag] and [TextFlag],
// a sensitive flag parses into a shallow copy of its default value rather
// than into the default value itself, so the object given as the default
// value keeps its default. Read the parsed value with methods such as
// [Command.Generic] and [Command.Text], or through Destination. As the copy
// is shallow, a value whose Set method writes into a map, slice or pointer
// that it shares with the default value also changes the default value.
type SensitiveFlag interface {
	// IsSensitive returns true if the flag's value must be kept secret
	IsSensitive() bool
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
	// TypeName to detect if a flag is a string, bool, etc.
	TypeName() string
}

// DocGenerationMultiValueFlag extends DocGenerationFlag for slice/map based flags.
type DocGenerationMultiValueFlag interface {
	DocGenerationFlag

	// IsMultiValueFlag returns true for flags that can be given multiple times.
	IsMultiValueFlag() bool
}

// SchemaTyper is an optional interface for flags that can report their
// JSON Schema type for programmatic introspection.
type SchemaTyper interface {
	// SchemaType returns the JSON Schema type name for the value this
	// flag accepts: "boolean", "integer", "number", "string", "array",
	// "object". Returns "" if the flag does not map cleanly.
	SchemaType() string
}

// SchemaItemsTyper is an optional interface for multi-value flags that
// can report the JSON Schema type of their elements.
type SchemaItemsTyper interface {
	// SchemaItemsType returns the JSON Schema type of elements for
	// array-type flags. Returns "" for single-value or object flags.
	SchemaItemsType() string
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

// StringerSetter is an optional interface that allows an individual
// flag to be given a per-flag override of [FlagStringer]. FlagBase and
// BoolWithInverseFlag implement this. It's used by
// [MutuallyExclusiveFlags.Stringer] to customize how flags within a
// mutually exclusive group are displayed in help output.
type StringerSetter interface {
	// SetStringer overrides the [FlagStringFunc] used by this flag's
	// String method. Passing nil restores the default behavior of using
	// the package-level [FlagStringer].
	SetStringer(FlagStringFunc)
}

// LocalFlag is an interface to enable detection of flags which are local
// to current command
type LocalFlag interface {
	IsLocal() bool
}

// redactedText replaces the value of a sensitive flag in output
const redactedText = "REDACTED"

func isSensitive(f Flag) bool {
	sf, ok := f.(SensitiveFlag)
	return ok && sf.IsSensitive()
}

// sensitivePartser is implemented by flags that can split a raw value into
// the parts that an error message may quote, such as the items of a slice
// or the values of a map
type sensitivePartser interface {
	sensitiveParts(val string) []string
}

// redactValue replaces each occurrence of val, and of each part of val that
// an error message may quote on its own, in s if f is sensitive
func redactValue(f Flag, s, val string) string {
	if !isSensitive(f) || val == "" {
		return s
	}

	parts := []string{val}
	if p, ok := f.(sensitivePartser); ok {
		parts = p.sensitiveParts(val)
	}

	// Replace longer parts first, so that a part that contains another is
	// replaced as a whole
	parts = slices.DeleteFunc(slices.Clone(parts), func(p string) bool { return p == "" })
	slices.SortStableFunc(parts, func(a, b string) int { return len(b) - len(a) })
	parts = slices.Compact(parts)

	oldnew := make([]string, 0, 2*len(parts))
	for _, p := range parts {
		oldnew = append(oldnew, p, redactedText)
	}
	return strings.NewReplacer(oldnew...).Replace(s)
}

// sensitiveStrings returns the string forms of a sensitive flag's value.
// Only values that are text are returned, as replacing e.g. a number
// everywhere in a log line would hide much more than the secret.
func sensitiveStrings(v any) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case map[string]string:
		vals := make([]string, 0, len(v))
		for _, val := range v {
			vals = append(vals, val)
		}
		return vals
	case fmt.Stringer:
		return []string{v.String()}
	}
	return nil
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

func FlagNames(name string, aliases []string) []string {
	var ret []string

	for _, part := range append([]string{name}, aliases...) {
		// v1 -> v2 migration warning zone:
		// Strip off anything after the first found comma or space, which
		// *hopefully* makes it a tiny bit more obvious that unexpected behavior is
		// caused by using the v1 form of stringly typed "Name".
		if i := strings.IndexAny(part, ", "); i >= 0 {
			ret = append(ret, part[:i])
		} else {
			ret = append(ret, part)
		}
	}

	return ret
}

func hasFlag(flags []Flag, fl Flag) bool {
	return slices.Contains(flags, fl)
}

func flagSplitMultiValues(val string, sliceSeparator string, disableSliceSeparator bool) []string {
	if disableSliceSeparator {
		return []string{val}
	}

	if len(sliceSeparator) == 0 {
		sliceSeparator = defaultSliceFlagSeparator
	}
	return strings.Split(val, sliceSeparator)
}
