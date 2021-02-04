package cli

import (
	"flag"
	"fmt"
	"os"
)

// EnumByName Returns the Enum by name from cli.Context.
func EnumByName(c *Context, name string) Enum {
	v := c.Value(name)
	h := v.(EnumHolder)
	return h.Value()
}

// Enum Defines an Enum.
type Enum interface {
	fmt.Stringer
}

// EnumDecoder A decoder that decodes to Enum.
type EnumDecoder interface {
	FromString(s string) (Enum, error)
	Strings() []string
}

// NewEnumHolder Initializes a new instance of EnumHolder.
func NewEnumHolder(value Enum, decoder EnumDecoder) *EnumHolder {
	return &EnumHolder{
		value:   value,
		decoder: decoder,
	}
}

// EnumHolder Holds an Enum value.
type EnumHolder struct {
	value      Enum
	decoder    EnumDecoder
	hasBeenSet bool
}

// String Returns the string representation of the Enum it holds.
func (h *EnumHolder) String() string {
	if h.value == nil {
		return "unsupported"
	}
	return h.value.String()
}

// Set the Enum based on its string representation.
func (h *EnumHolder) Set(s string) (err error) {
	if h.value, err = h.decoder.FromString(s); err != nil {
		return err
	}
	h.hasBeenSet = true
	return nil
}

// Value returns the Enum.
func (h *EnumHolder) Value() Enum {
	return h.value
}

// Get Returns a copy of JETModeHolder.
func (h EnumHolder) Get() interface{} {
	return h
}

// EnumFlag A cli Flag that holds a Enum.
type EnumFlag struct {
	Name        string
	Aliases     []string
	Value       *EnumHolder
	Decoder     EnumDecoder
	EnvVars     []string
	Usage       string
	DefaultText string
	Required    bool
	HasBeenSet  bool
}

// String Describes the Flag to the Caller.
func (f *EnumFlag) String() string {
	return fmt.Sprintf("%s (supported values: %s)", FlagStringer(f), f.Decoder.Strings())
}

// Apply the value of the Flag to the cli.
func (f *EnumFlag) Apply(set *flag.FlagSet) error {
	if v, ok := stringFromEnvs(f.EnvVars); ok {
		v, err := f.Decoder.FromString(v)
		if err != nil {
			return fmt.Errorf("supported values: %s", f.Decoder.Strings())
		}
		f.Value = NewEnumHolder(v, f.Decoder)
		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		set.Var(f.Value, name, f.Usage)
	}
	if f.DefaultText == "" {
		f.DefaultText = f.Value.String()
	}
	return nil
}

// Names Returns all flag names of this cli.Flag.
func (f *EnumFlag) Names() []string {
	return append(f.Aliases, f.Name)
}

// IsSet Whether this cli.Flag has been set or not.
func (f *EnumFlag) IsSet() bool {
	return f.HasBeenSet
}

// IsRequired Whether this cli.Flag is required or not.
func (f *EnumFlag) IsRequired() bool {
	return f.Required
}

// TakesValue Whether this cli.Flag takes a value or not.
func (f *EnumFlag) TakesValue() bool {
	return true
}

// GetUsage Returns the usage description of this cli.Flag.
func (f *EnumFlag) GetUsage() string {
	return f.Usage
}

// GetValue Returns the current value of this cli.Flag.
func (f *EnumFlag) GetValue() string {
	return f.Value.String()
}

func stringFromEnvs(vars []string) (string, bool) {
	for _, env := range vars {
		if v, ok := os.LookupEnv(env); ok {
			return v, true
		}
	}
	return "", false
}
