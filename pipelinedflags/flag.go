package altinputsource

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

// FlagInputSourceExtension has extensions that allow the
// pipelined version of FlagSetWrappers to be used effectively
type FlagInputSourceExtension interface {
	getEnvVar() string
	getDefaultValue() interface{}
}

// GenericFlag is the flag type for types implementing Generic
type GenericFlag struct {
	Name   string
	Value  cli.Generic
	Usage  string
	EnvVar string
}

// String returns the string representation of the generic flag to display the
// help text to the user (uses the String() method of the generic flag to show
// the value)
func (f GenericFlag) String() string {
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s%s \"%v\"\t%v", prefixFor(f.Name), f.Name, f.Value, f.Usage))
}

// Apply takes the flagset and calls Set on the generic flag with the value
// provided by the user for parsing by the flag
func (f GenericFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Var(f.Value, name, f.Usage)
	})
}

func (f GenericFlag) GetName() string {
	return f.Name
}

func (f GenericFlag) getEnvVar() string {
	return f.EnvVar
}

func (f GenericFlag) getDefaultValue() interface{} {
	return f.Value
}

// StringSlice is a string flag that can be specified multiple times on the
// command-line
type StringSliceFlag struct {
	Name   string
	Value  *cli.StringSlice
	Usage  string
	EnvVar string
}

// String returns the usage
func (f StringSliceFlag) String() string {
	firstName := strings.Trim(strings.Split(f.Name, ",")[0], " ")
	pref := prefixFor(firstName)
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s [%v]\t%v", prefixedNames(f.Name), pref+firstName+" option "+pref+firstName+" option", f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f StringSliceFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Var(&cli.StringSlice{}, name, f.Usage)
	})
}

func (f StringSliceFlag) GetName() string {
	return f.Name
}

func (f StringSliceFlag) getEnvVar() string {
	return f.EnvVar
}

func (f StringSliceFlag) getDefaultValue() interface{} {
	return f.Value
}

// IntSliceFlag is an int flag that can be specified multiple times on the
// command-line
type IntSliceFlag struct {
	Name   string
	Value  *cli.IntSlice
	Usage  string
	EnvVar string
}

// String returns the usage
func (f IntSliceFlag) String() string {
	firstName := strings.Trim(strings.Split(f.Name, ",")[0], " ")
	pref := prefixFor(firstName)
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s [%v]\t%v", prefixedNames(f.Name), pref+firstName+" option "+pref+firstName+" option", f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f IntSliceFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Var(&cli.IntSlice{}, name, f.Usage)
	})
}

func (f IntSliceFlag) GetName() string {
	return f.Name
}

func (f IntSliceFlag) getEnvVar() string {
	return f.EnvVar
}

func (f IntSliceFlag) getDefaultValue() interface{} {
	return f.Value
}

// BoolFlag is a switch that defaults to false
type BoolFlag struct {
	Name        string
	Usage       string
	EnvVar      string
	Destination *bool
}

// String returns a readable representation of this value (for usage defaults)
func (f BoolFlag) String() string {
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s\t%v", prefixedNames(f.Name), f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f BoolFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Bool(name, false, f.Usage)
	})
}

func (f BoolFlag) GetName() string {
	return f.Name
}

func (f BoolFlag) getEnvVar() string {
	return f.EnvVar
}

func (f BoolFlag) getDefaultValue() interface{} {
	return *f.Destination
}

// BoolTFlag this represents a boolean flag that is true by default, but can
// still be set to false by --some-flag=false
type BoolTFlag struct {
	Name        string
	Usage       string
	EnvVar      string
	Destination *bool
}

// String returns a readable representation of this value (for usage defaults)
func (f BoolTFlag) String() string {
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s\t%v", prefixedNames(f.Name), f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f BoolTFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Bool(name, true, f.Usage)
	})
}

func (f BoolTFlag) GetName() string {
	return f.Name
}

func (f BoolTFlag) getEnvVar() string {
	return f.EnvVar
}

func (f BoolTFlag) getDefaultValue() interface{} {
	return *f.Destination
}

// StringFlag represents a flag that takes as string value
type StringFlag struct {
	Name        string
	Value       string
	Usage       string
	EnvVar      string
	Destination *string
}

// String returns the usage
func (f StringFlag) String() string {
	var fmtString string
	fmtString = "%s %v\t%v"

	if len(f.Value) > 0 {
		fmtString = "%s \"%v\"\t%v"
	} else {
		fmtString = "%s %v\t%v"
	}

	return withEnvHint(f.EnvVar, fmt.Sprintf(fmtString, prefixedNames(f.Name), f.Value, f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f StringFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.String(name, "", f.Usage)
	})
}

func (f StringFlag) GetName() string {
	return f.Name
}

func (f StringFlag) getEnvVar() string {
	return f.EnvVar
}

func (f StringFlag) getDefaultValue() interface{} {
	return *f.Destination
}

// IntFlag is a flag that takes an integer
// Errors if the value provided cannot be parsed
type IntFlag struct {
	Name        string
	Value       int
	Usage       string
	EnvVar      string
	Destination *int
}

// String returns the usage
func (f IntFlag) String() string {
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s \"%v\"\t%v", prefixedNames(f.Name), f.Value, f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f IntFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Int(name, 0, f.Usage)
	})
}

func (f IntFlag) GetName() string {
	return f.Name
}

func (f IntFlag) getEnvVar() string {
	return f.EnvVar
}

func (f IntFlag) getDefaultValue() interface{} {
	return f.Value
}

// DurationFlag is a flag that takes a duration specified in Go's duration
// format: https://golang.org/pkg/time/#ParseDuration
type DurationFlag struct {
	Name        string
	Value       time.Duration
	Usage       string
	EnvVar      string
	Destination *time.Duration
}

// String returns a readable representation of this value (for usage defaults)
func (f DurationFlag) String() string {
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s \"%v\"\t%v", prefixedNames(f.Name), f.Value, f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f DurationFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Duration(name, 0, f.Usage)
	})
}

func (f DurationFlag) getName() string {
	return f.Name
}

func (f DurationFlag) getEnvVar() string {
	return f.EnvVar
}

func (f DurationFlag) getDefaultValue() interface{} {
	return f.Value
}

// Float64Flag is a flag that takes an float value
// Errors if the value provided cannot be parsed
type Float64Flag struct {
	Name        string
	Value       float64
	Usage       string
	EnvVar      string
	Destination *float64
}

// String returns the usage
func (f Float64Flag) String() string {
	return withEnvHint(f.EnvVar, fmt.Sprintf("%s \"%v\"\t%v", prefixedNames(f.Name), f.Value, f.Usage))
}

// Apply populates the flag given the flag set and environment
func (f Float64Flag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Float64(name, 0, f.Usage)
	})
}

func (f Float64Flag) getName() string {
	return f.Name
}

func (f Float64Flag) getEnvVar() string {
	return f.EnvVar
}

func (f Float64Flag) getDefaultValue() interface{} {
	return f.Value
}

// TODO: Copied code below, should avoid duplication
func eachName(longName string, fn func(string)) {
	parts := strings.Split(longName, ",")
	for _, name := range parts {
		name = strings.Trim(name, " ")
		fn(name)
	}
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}

func prefixedNames(fullName string) (prefixed string) {
	parts := strings.Split(fullName, ",")
	for i, name := range parts {
		name = strings.Trim(name, " ")
		prefixed += prefixFor(name) + name
		if i < len(parts)-1 {
			prefixed += ", "
		}
	}
	return
}

func withEnvHint(envVar, str string) string {
	envText := ""
	if envVar != "" {
		envText = fmt.Sprintf(" [$%s]", strings.Join(strings.Split(envVar, ","), ", $"))
	}
	return str + envText
}
