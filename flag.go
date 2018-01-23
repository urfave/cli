package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const defaultPlaceholder = "value"

var (
	slPfx = fmt.Sprintf("sl:::%d:::", time.Now().UTC().UnixNano())

	commaWhitespace = regexp.MustCompile("[, ]+.*")
)

// GenerateCompletionFlag enables completion for all commands and subcommands
var GenerateCompletionFlag Flag = &BoolFlag{
	Name:   "generate-completion",
	Hidden: true,
}

func genCompName() string {
	names := GenerateCompletionFlag.Names()
	if len(names) == 0 {
		return "generate-completion"
	}
	return names[0]
}

// InitCompletionFlag generates completion code
var InitCompletionFlag = &StringFlag{
	Name:  "init-completion",
	Usage: "generate completion code. Value must be 'bash' or 'zsh'",
}

// VersionFlag prints the version for the application
var VersionFlag Flag = &BoolFlag{
	Name:    "version",
	Aliases: []string{"v"},
	Usage:   "print the version",
}

// HelpFlag prints the help for all commands and subcommands.
// Set to nil to disable the flag.  The subcommand
// will still be added unless HideHelp is set to true.
var HelpFlag Flag = &BoolFlag{
	Name:    "help",
	Aliases: []string{"h"},
	Usage:   "show help",
}

// FlagStringer converts a flag definition to a string. This is used by help
// to display a flag.
var FlagStringer FlagStringFunc = stringifyFlag

// Serializeder is used to circumvent the limitations of flag.FlagSet.Set
type Serializeder interface {
	Serialized() string
}

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
	return f[i].Names()[0] < f[j].Names()[0]
}

func (f FlagsByName) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// Flag is a common interface related to parsing flags in cli.
// For more advanced flag parsing techniques, it is recommended that
// this interface be implemented.
type Flag interface {
	fmt.Stringer
	// Apply Flag settings to the given flag set
	Apply(*flag.FlagSet)
	Names() []string
}

// errorableFlag is an interface that allows us to return errors during apply
// it allows flags defined in this library to return errors in a fashion backwards compatible
// TODO remove in v2 and modify the existing Flag interface to return errors
type errorableFlag interface {
	Flag

	ApplyWithError(*flag.FlagSet) error
}

func flagSet(name string, flags []Flag) (*flag.FlagSet, error) {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		//TODO remove in v2 when errorableFlag is removed
		if ef, ok := f.(errorableFlag); ok {
			if err := ef.ApplyWithError(set); err != nil {
				return nil, err
			}
		} else {
			f.Apply(set)
		}
	}
	return set, nil
}

// Generic is a generic parseable type identified by a specific flag
type Generic interface {
	Set(value string) error
	String() string
}

// Apply takes the flagset and calls Set on the generic flag with the value
// provided by the user for parsing by the flag
// Ignores parsing errors
func (f *GenericFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError takes the flagset and calls Set on the generic flag with the
// value provided by the user for parsing by the flag
func (f *GenericFlag) ApplyWithError(set *flag.FlagSet) error {
	val := f.Value
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				val.Set(envVal)
				break
			}
		}
	}

	for _, name := range f.Names() {
		set.Var(val, name, f.Usage)
	}
	return nil
}

// StringSlice wraps a []string to satisfy flag.Value
type StringSlice struct {
	slice      []string
	hasBeenSet bool
}

// NewStringSlice creates a *StringSlice with default values
func NewStringSlice(defaults ...string) *StringSlice {
	return &StringSlice{slice: append([]string{}, defaults...)}
}

// Set appends the string value to the list of values
func (f *StringSlice) Set(value string) error {
	if !f.hasBeenSet {
		f.slice = []string{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &f.slice)
		f.hasBeenSet = true
		return nil
	}

	f.slice = append(f.slice, value)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *StringSlice) String() string {
	return fmt.Sprintf("%s", f.slice)
}

// Serialized allows StringSlice to fulfill Serializeder
func (f *StringSlice) Serialized() string {
	jsonBytes, _ := json.Marshal(f.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of strings set by this flag
func (f *StringSlice) Value() []string {
	return f.slice
}

// Get returns the slice of strings set by this flag
func (f *StringSlice) Get() interface{} {
	return *f
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *StringSliceFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *StringSliceFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				newVal := NewStringSlice()
				for _, s := range strings.Split(envVal, ",") {
					s = strings.TrimSpace(s)
					if err := newVal.Set(s); err != nil {
						return fmt.Errorf("could not parse %q as string value for flag %s: %s", envVal, f.Name, err)
					}
				}
				f.Value = newVal
				break
			}
		}
	}

	if f.Value == nil {
		f.Value = NewStringSlice()
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}
	return nil
}

// IntSlice wraps an []int to satisfy flag.Value
type IntSlice struct {
	slice      []int
	hasBeenSet bool
}

// NewIntSlice makes an *IntSlice with default values
func NewIntSlice(defaults ...int) *IntSlice {
	return &IntSlice{slice: append([]int{}, defaults...)}
}

// NewInt64Slice makes an *Int64Slice with default values
func NewInt64Slice(defaults ...int64) *Int64Slice {
	return &Int64Slice{slice: append([]int64{}, defaults...)}
}

// SetInt directly adds an integer to the list of values
func (i *IntSlice) SetInt(value int) {
	if !i.hasBeenSet {
		i.slice = []int{}
		i.hasBeenSet = true
	}

	i.slice = append(i.slice, value)
}

// Set parses the value into an integer and appends it to the list of values
func (i *IntSlice) Set(value string) error {
	if !i.hasBeenSet {
		i.slice = []int{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &i.slice)
		i.hasBeenSet = true
		return nil
	}

	tmp, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return err
	}

	i.slice = append(i.slice, int(tmp))
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *IntSlice) String() string {
	return fmt.Sprintf("%#v", f.slice)
}

// Serialized allows IntSlice to fulfill Serializeder
func (i *IntSlice) Serialized() string {
	jsonBytes, _ := json.Marshal(i.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (i *IntSlice) Value() []int {
	return i.slice
}

// Get returns the slice of ints set by this flag
func (f *IntSlice) Get() interface{} {
	return *f
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *IntSliceFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *IntSliceFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				newVal := NewIntSlice()
				for _, s := range strings.Split(envVal, ",") {
					s = strings.TrimSpace(s)
					if err := newVal.Set(s); err != nil {
						return fmt.Errorf("could not parse %q as int slice value for flag %s: %s", envVal, f.Name, err)
					}
				}
				f.Value = newVal
				break
			}
		}
	}

	if f.Value == nil {
		f.Value = NewIntSlice()
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}
	return nil
}

// Int64Slice is an opaque type for []int to satisfy flag.Value
type Int64Slice struct {
	slice      []int64
	hasBeenSet bool
}

// Set parses the value into an integer and appends it to the list of values
func (f *Int64Slice) Set(value string) error {
	if !f.hasBeenSet {
		f.slice = []int64{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &f.slice)
		f.hasBeenSet = true
		return nil
	}

	tmp, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return err
	}

	f.slice = append(f.slice, tmp)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *Int64Slice) String() string {
	return fmt.Sprintf("%#v", f.slice)
}

// Serialized allows Int64Slice to fulfill Serializeder
func (f *Int64Slice) Serialized() string {
	jsonBytes, _ := json.Marshal(f.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (f *Int64Slice) Value() []int64 {
	return f.slice
}

// Get returns the slice of ints set by this flag
func (f *Int64Slice) Get() interface{} {
	return *f
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *Int64SliceFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *Int64SliceFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				newVal := NewInt64Slice()
				for _, s := range strings.Split(envVal, ",") {
					s = strings.TrimSpace(s)
					if err := newVal.Set(s); err != nil {
						return fmt.Errorf("could not parse %q as int64 slice value for flag %s: %s", envVal, f.Name, err)
					}
				}
				f.Value = newVal
				break
			}
		}
	}

	if f.Value == nil {
		f.Value = NewInt64Slice()
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *BoolFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *BoolFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				if envVal == "" {
					f.Value = false
					break
				}

				envValBool, err := strconv.ParseBool(envVal)
				if err != nil {
					return fmt.Errorf("could not parse %q as bool value for flag %s: %s", envVal, f.Name, err)
				}
				f.Value = envValBool
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.BoolVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Bool(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *StringFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *StringFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				f.Value = envVal
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.StringVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.String(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *PathFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *PathFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				f.Value = envVal
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.StringVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.String(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *IntFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *IntFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				envValInt, err := strconv.ParseInt(envVal, 0, 64)
				if err != nil {
					return fmt.Errorf("could not parse %q as int value for flag %s: %s", envVal, f.Name, err)
				}
				f.Value = int(envValInt)
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.IntVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Int(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *Int64Flag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *Int64Flag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				envValInt, err := strconv.ParseInt(envVal, 0, 64)
				if err != nil {
					return fmt.Errorf("could not parse %q as int value for flag %s: %s", envVal, f.Name, err)
				}

				f.Value = envValInt
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Int64Var(f.Destination, name, f.Value, f.Usage)
			return nil
		}
		set.Int64(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *UintFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *UintFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				envValInt, err := strconv.ParseUint(envVal, 0, 64)
				if err != nil {
					return fmt.Errorf("could not parse %q as uint value for flag %s: %s", envVal, f.Name, err)
				}

				f.Value = uint(envValInt)
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.UintVar(f.Destination, name, f.Value, f.Usage)
			return nil
		}
		set.Uint(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *Uint64Flag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *Uint64Flag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				envValInt, err := strconv.ParseUint(envVal, 0, 64)
				if err != nil {
					return fmt.Errorf("could not parse %q as uint64 value for flag %s: %s", envVal, f.Name, err)
				}

				f.Value = uint64(envValInt)
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Uint64Var(f.Destination, name, f.Value, f.Usage)
			return nil
		}
		set.Uint64(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *DurationFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *DurationFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				envValDuration, err := time.ParseDuration(envVal)
				if err != nil {
					return fmt.Errorf("could not parse %q as duration for flag %s: %s", envVal, f.Name, err)
				}

				f.Value = envValDuration
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.DurationVar(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Duration(name, f.Value, f.Usage)
	}
	return nil
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *Float64Flag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *Float64Flag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				envValFloat, err := strconv.ParseFloat(envVal, 10)
				if err != nil {
					return fmt.Errorf("could not parse %q as float64 value for flag %s: %s", envVal, f.Name, err)
				}

				f.Value = float64(envValFloat)
				break
			}
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Float64Var(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Float64(name, f.Value, f.Usage)
	}
	return nil
}

// NewFloat64Slice makes a *Float64Slice with default values
func NewFloat64Slice(defaults ...float64) *Float64Slice {
	return &Float64Slice{slice: append([]float64{}, defaults...)}
}

// Float64Slice is an opaque type for []float64 to satisfy flag.Value
type Float64Slice struct {
	slice      []float64
	hasBeenSet bool
}

// Set parses the value into a float64 and appends it to the list of values
func (f *Float64Slice) Set(value string) error {
	if !f.hasBeenSet {
		f.slice = []float64{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &f.slice)
		f.hasBeenSet = true
		return nil
	}

	tmp, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	f.slice = append(f.slice, tmp)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *Float64Slice) String() string {
	return fmt.Sprintf("%#v", f.slice)
}

// Serialized allows Float64Slice to fulfill Serializeder
func (f *Float64Slice) Serialized() string {
	jsonBytes, _ := json.Marshal(f.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of float64s set by this flag
func (f *Float64Slice) Value() []float64 {
	return f.slice
}

// Apply populates the flag given the flag set and environment
// Ignores errors
func (f *Float64SliceFlag) Apply(set *flag.FlagSet) {
	f.ApplyWithError(set)
}

// ApplyWithError populates the flag given the flag set and environment
func (f *Float64SliceFlag) ApplyWithError(set *flag.FlagSet) error {
	if f.EnvVars != nil {
		for _, envVar := range f.EnvVars {
			if envVal, ok := syscall.Getenv(envVar); ok {
				newVal := NewFloat64Slice()
				for _, s := range strings.Split(envVal, ",") {
					s = strings.TrimSpace(s)
					err := newVal.Set(s)
					if err != nil {
						fmt.Fprintf(ErrWriter, err.Error())
					}
				}
				f.Value = newVal
				break
			}
		}
	}

	if f.Value == nil {
		f.Value = NewFloat64Slice()
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}
	return nil
}

func visibleFlags(fl []Flag) []Flag {
	visible := []Flag{}
	for _, flag := range fl {
		field := flagValue(flag).FieldByName("Hidden")
		if !field.IsValid() || !field.Bool() {
			visible = append(visible, flag)
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

func withEnvHint(envVars []string, str string) string {
	envText := ""
	if envVars != nil && len(envVars) > 0 {
		prefix := "$"
		suffix := ""
		sep := ", $"
		if runtime.GOOS == "windows" {
			prefix = "%"
			suffix = "%"
			sep = "%, %"
		}
		envText = fmt.Sprintf(" [%s%s%s]", prefix, strings.Join(envVars, sep), suffix)
	}
	return str + envText
}

func flagNames(f Flag) []string {
	ret := []string{}

	name := flagStringField(f, "Name")
	aliases := flagStringSliceField(f, "Aliases")

	for _, part := range append([]string{name}, aliases...) {
		// v1 -> v2 migration warning zone:
		// Strip off anything after the first found comma or space, which
		// *hopefully* makes it a tiny bit more obvious that unexpected behavior is
		// caused by using the v1 form of stringly typed "Name".
		ret = append(ret, commaWhitespace.ReplaceAllString(part, ""))
	}

	return ret
}

func flagStringSliceField(f Flag, name string) []string {
	fv := flagValue(f)
	field := fv.FieldByName(name)

	if field.IsValid() {
		return field.Interface().([]string)
	}

	return []string{}
}

func flagStringField(f Flag, name string) string {
	fv := flagValue(f)
	field := fv.FieldByName(name)

	if field.IsValid() {
		return field.String()
	}

	return ""
}

func flagValue(f Flag) reflect.Value {
	fv := reflect.ValueOf(f)
	for fv.Kind() == reflect.Ptr {
		fv = reflect.Indirect(fv)
	}
	return fv
}

func stringifyFlag(f Flag) string {
	fv := flagValue(f)

	switch f.(type) {
	case *IntSliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyIntSliceFlag(f.(*IntSliceFlag)))
	case *Int64SliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyInt64SliceFlag(f.(*Int64SliceFlag)))
	case *Float64SliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyFloat64SliceFlag(f.(*Float64SliceFlag)))
	case *StringSliceFlag:
		return withEnvHint(flagStringSliceField(f, "EnvVars"),
			stringifyStringSliceFlag(f.(*StringSliceFlag)))
	}

	placeholder, usage := unquoteUsage(fv.FieldByName("Usage").String())

	needsPlaceholder := false
	defaultValueString := ""
	val := fv.FieldByName("Value")
	if val.IsValid() {
		needsPlaceholder = val.Kind() != reflect.Bool
		defaultValueString = fmt.Sprintf(" (default: %v)", val.Interface())

		if val.Kind() == reflect.String && val.String() != "" {
			defaultValueString = fmt.Sprintf(" (default: %q)", val.String())
		}
	}

	helpText := fv.FieldByName("DefaultText")
	if helpText.IsValid() && helpText.String() != "" {
		needsPlaceholder = val.Kind() != reflect.Bool
		defaultValueString = fmt.Sprintf(" (default: %s)", helpText.String())
	}

	if defaultValueString == " (default: )" {
		defaultValueString = ""
	}

	if needsPlaceholder && placeholder == "" {
		placeholder = defaultPlaceholder
	}

	usageWithDefault := strings.TrimSpace(fmt.Sprintf("%s%s", usage, defaultValueString))

	return withEnvHint(flagStringSliceField(f, "EnvVars"),
		fmt.Sprintf("%s\t%s", prefixedNames(f.Names(), placeholder), usageWithDefault))
}

func stringifyIntSliceFlag(f *IntSliceFlag) string {
	defaultVals := []string{}
	if f.Value != nil && len(f.Value.Value()) > 0 {
		for _, i := range f.Value.Value() {
			defaultVals = append(defaultVals, fmt.Sprintf("%d", i))
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifyInt64SliceFlag(f *Int64SliceFlag) string {
	defaultVals := []string{}
	if f.Value != nil && len(f.Value.Value()) > 0 {
		for _, i := range f.Value.Value() {
			defaultVals = append(defaultVals, fmt.Sprintf("%d", i))
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifyFloat64SliceFlag(f *Float64SliceFlag) string {
	defaultVals := []string{}
	if f.Value != nil && len(f.Value.Value()) > 0 {
		for _, i := range f.Value.Value() {
			defaultVals = append(defaultVals, strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", i), "0"), "."))
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifyStringSliceFlag(f *StringSliceFlag) string {
	defaultVals := []string{}
	if f.Value != nil && len(f.Value.Value()) > 0 {
		for _, s := range f.Value.Value() {
			if len(s) > 0 {
				defaultVals = append(defaultVals, fmt.Sprintf("%q", s))
			}
		}
	}

	return stringifySliceFlag(f.Usage, f.Names(), defaultVals)
}

func stringifySliceFlag(usage string, names, defaultVals []string) string {
	placeholder, usage := unquoteUsage(usage)
	if placeholder == "" {
		placeholder = defaultPlaceholder
	}

	defaultVal := ""
	if len(defaultVals) > 0 {
		defaultVal = fmt.Sprintf(" (default: %s)", strings.Join(defaultVals, ", "))
	}

	usageWithDefault := strings.TrimSpace(fmt.Sprintf("%s%s", usage, defaultVal))
	return fmt.Sprintf("%s\t%s", prefixedNames(names, placeholder), usageWithDefault)
}

func hasFlag(flags []Flag, fl Flag) bool {
	for _, existing := range flags {
		if fl == existing {
			return true
		}
	}

	return false
}
