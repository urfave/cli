package cli

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type (

	// Flag is a common interface related to parsing flags in cli.
	// For more advanced flag parsing techniques, it is recomended that
	// this interface be implemented.
	Flag interface {
		fmt.Stringer
		// Apply Flag settings to the given flag set
		Apply(*flag.FlagSet)
		getName() string
	}

	StringSlice []string

	StringSliceFlag struct {
		Name  string
		Value *StringSlice
		Usage string
	}

	IntSlice []int

	IntSliceFlag struct {
		Name  string
		Value *IntSlice
		Usage string
	}

	BoolFlag struct {
		Name  string
		Usage string
	}

	// Same structure
	BoolTFlag BoolFlag

	StringFlag struct {
		Name  string
		Value string
		Usage string
	}

	IntFlag struct {
		Name  string
		Value int
		Usage string
	}

	Float64Flag struct {
		Name  string
		Value float64
		Usage string
	}
)

// This flag enables bash-completion for all commands and subcommands
var BashCompletionFlag = BoolFlag{"generate-bash-completion", ""}

// Utility functions

func flagSet(name string, flags []Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

func eachName(longName string, fn func(string)) {
	parts := strings.Split(longName, ",")
	for _, name := range parts {
		name = strings.Trim(name, " ")
		fn(name)
	}
}

func prefixFor(name string) string {
	if len(name) == 1 {
		return "-"
	}
	return "--"
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

// --- StringSlice ---

func (f *StringSlice) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *StringSlice) String() string {
	return fmt.Sprintf("%s", *f)
}

func (f *StringSlice) Value() []string {
	return *f
}

// --- StringSliceFlag ---

func (f StringSliceFlag) String() string {
	firstName := strings.Trim(strings.Split(f.Name, ",")[0], " ")
	pref := prefixFor(firstName)
	return fmt.Sprintf("%s '%v'\t%v", prefixedNames(f.Name), pref+firstName+" option "+pref+firstName+" option", f.Usage)
}

func (f StringSliceFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Var(f.Value, name, f.Usage)
	})
}

func (f StringSliceFlag) getName() string {
	return f.Name
}

// --- IntSlice ---

func (f *IntSlice) Set(value string) error {

	tmp, err := strconv.Atoi(value)
	if err != nil {
		return err
	} else {
		*f = append(*f, tmp)
	}
	return nil
}

func (f *IntSlice) String() string {
	return fmt.Sprintf("%d", *f)
}

func (f *IntSlice) Value() []int {
	return *f
}

// --- IntSliceFlag ---

func (f IntSliceFlag) String() string {
	firstName := strings.Trim(strings.Split(f.Name, ",")[0], " ")
	pref := prefixFor(firstName)
	return fmt.Sprintf("%s '%v'\t%v", prefixedNames(f.Name), pref+firstName+" option "+pref+firstName+" option", f.Usage)
}

func (f IntSliceFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Var(f.Value, name, f.Usage)
	})
}

func (f IntSliceFlag) getName() string {
	return f.Name
}

// --- BoolFlag ---

func (f BoolFlag) String() string {
	return fmt.Sprintf("%s\t%v", prefixedNames(f.Name), f.Usage)
}

func (f BoolFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Bool(name, false, f.Usage)
	})
}

func (f BoolFlag) getName() string {
	return f.Name
}

// --- BoolTFlag ---

func (f BoolTFlag) String() string {
	return fmt.Sprintf("%s\t%v", prefixedNames(f.Name), f.Usage)
}

func (f BoolTFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Bool(name, true, f.Usage)
	})
}

func (f BoolTFlag) getName() string {
	return f.Name
}

// --- StringFlag ---

func (f StringFlag) String() string {
	var fmtString string
	fmtString = "%s %v\t%v"

	if len(f.Value) > 0 {
		fmtString = "%s '%v'\t%v"
	} else {
		fmtString = "%s %v\t%v"
	}

	return fmt.Sprintf(fmtString, prefixedNames(f.Name), f.Value, f.Usage)
}

func (f StringFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.String(name, f.Value, f.Usage)
	})
}

func (f StringFlag) getName() string {
	return f.Name
}

// --- IntFlag ---

func (f IntFlag) String() string {
	return fmt.Sprintf("%s '%v'\t%v", prefixedNames(f.Name), f.Value, f.Usage)
}

func (f IntFlag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Int(name, f.Value, f.Usage)
	})
}

func (f IntFlag) getName() string {
	return f.Name
}

// --- Float64Flag ---

func (f Float64Flag) String() string {
	return fmt.Sprintf("%s '%v'\t%v", prefixedNames(f.Name), f.Value, f.Usage)
}

func (f Float64Flag) Apply(set *flag.FlagSet) {
	eachName(f.Name, func(name string) {
		set.Float64(name, f.Value, f.Usage)
	})
}

func (f Float64Flag) getName() string {
	return f.Name
}
