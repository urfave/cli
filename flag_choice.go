package cli

import (
	"errors"
	"flag"
	"fmt"
)

var errParse = errors.New("parse error")

// Choice Defines the definition of a choice.
type Choice interface {
	// FromString Returns a Choice value from string.
	FromString(s string) interface{}

	// ToString Returns the string representation of the given Choice value.
	ToString(i interface{}) string

	// Strings Returns all possible Choice values as string representation.
	Strings() []string
}

// NewStringChoiceDecoder Initializes a new instance of Choice that takes a list of strings used as choices.
func NewStringChoiceDecoder(ss ...string) Choice {
	c := make(map[string]interface{}, len(ss))
	for _, s := range ss {
		c[s] = s
	}
	return NewChoiceDecoder(c)
}

// Choices Maps a unique string value to any value.
type Choices map[string]interface{}

// NewChoiceDecoder Initializes a new default implementation of Choice.
// The provided Choices need to have unique values.
func NewChoiceDecoder(v Choices) Choice {
	out := new(defaultChoiceDecoder)
	out.init(v)
	return out
}

type defaultChoiceDecoder struct {
	vMap map[string]interface{}
	sMap map[interface{}]string
	ss   []string
}

func (d *defaultChoiceDecoder) init(v Choices) {
	d.vMap = v

	d.sMap = make(map[interface{}]string, len(v))
	for k, v := range v {
		d.sMap[v] = k
	}

	d.ss = make([]string, len(v))
	i := 0
	for k := range v {
		d.ss[i] = k
		i++
	}
}

func (d *defaultChoiceDecoder) FromString(s string) interface{} {
	if v, ok := d.vMap[s]; ok {
		return v
	}
	return nil
}

func (d *defaultChoiceDecoder) ToString(v interface{}) string {
	if v, ok := d.sMap[v]; ok {
		return v
	}
	return ""
}

func (d *defaultChoiceDecoder) Strings() []string {
	return d.ss
}

// ChoiceFlag A cli Flag that holds a Choice.
type ChoiceFlag struct {
	Name        string
	Aliases     []string
	Value       interface{}
	Decoder     Choice
	EnvVars     []string
	FilePath    string
	Usage       string
	DefaultText string
	Required    bool
	Destination *interface{}
	HasBeenSet  bool
}

// String Describes the Flag to the caller.
func (f *ChoiceFlag) String() string {
	return FlagStringer(f)
}

// Apply the value of the Flag to the cli.
func (f *ChoiceFlag) Apply(set *flag.FlagSet) error {
	if f.Decoder == nil {
		return fmt.Errorf("decoder must be provided for ChoiceFlag")
	}

	if v, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		v := f.Decoder.FromString(v)
		if v == nil {
			return errParse
		}
		f.Value = v
		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Var(newChoiceValueSwap(f.Decoder, f.Value, f.Destination), name, f.Usage)
			continue
		}
		set.Var(newChoiceValue(f.Decoder, f.Value), name, f.Usage)
	}

	return nil
}

// Names Returns all flag names of this cli.Flag.
func (f *ChoiceFlag) Names() []string {
	return append(f.Aliases, f.Name)
}

// IsSet Whether this cli.Flag has been set or not.
func (f *ChoiceFlag) IsSet() bool {
	return f.HasBeenSet
}

// IsRequired Whether this cli.Flag is required or not.
func (f *ChoiceFlag) IsRequired() bool {
	return f.Required
}

// TakesValue Whether this cli.Flag takes a value or not.
func (f *ChoiceFlag) TakesValue() bool {
	return true
}

// GetUsage Returns the usage description of this cli.Flag.
func (f *ChoiceFlag) GetUsage() string {
	return f.Usage
}

// GetValue Returns the current value of this cli.Flag.
func (f *ChoiceFlag) GetValue() string {
	return f.Decoder.ToString(f.Value)
}

// Choice looks up the value of a local ChoiceFlag.
// Returns nil if not found.
func (c *Context) Choice(name string) interface{} {
	v := c.Value(name)
	if h, ok := v.(choiceValue); ok {
		return h.Value()
	}
	return nil
}

type choiceValue struct {
	value   *interface{}
	decoder Choice
}

func newChoiceValue(decoder Choice, val interface{}) *choiceValue {
	return &choiceValue{decoder: decoder, value: &val}
}

func newChoiceValueSwap(decoder Choice, val interface{}, p *interface{}) *choiceValue {
	*p = val
	return &choiceValue{
		value:   p,
		decoder: decoder,
	}
}

func (c *choiceValue) Set(s string) error {
	v := c.decoder.FromString(s)
	*c.value = v
	if v == nil {
		return errParse
	}
	return nil
}

func (c choiceValue) Get() interface{} { return c }

func (c *choiceValue) String() string {
	if c.value == nil {
		return ""
	}
	if v := *c.value; v != nil {
		return c.decoder.ToString(v)
	}
	return ""
}

func (c *choiceValue) Value() interface{} { return *c.value }
