package cli

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

var errParse = errors.New("parse error")

// Choice Defines an Choice.
type Choice interface {
	fmt.Stringer
}

// ChoiceDecoder A decoder that decodes to Choice.
type ChoiceDecoder interface {
	FromString(s string) (Choice, error)
	Strings() []string
}

// NewChoiceHolder Initializes a new instance of ChoiceHolder.
func NewChoiceHolder(value Choice) *ChoiceHolder {
	return &ChoiceHolder{
		value: value,
	}
}

// ChoiceHolder Holds an Choice value.
type ChoiceHolder struct {
	value      Choice
	decoder    ChoiceDecoder
	hasBeenSet bool
}

func (h *ChoiceHolder) init(decoder ChoiceDecoder) {
	h.decoder = decoder
}

// String Returns the string representation of the Choice it holds.
func (h *ChoiceHolder) String() string {
	if h.value == nil {
		return ""
	}
	return h.value.String()
}

// Set the Choice based on its string representation.
func (h *ChoiceHolder) Set(s string) (err error) {
	if h.value, err = h.decoder.FromString(s); err != nil {
		return err
	}
	h.hasBeenSet = true
	return nil
}

// Value returns the Choice.
func (h *ChoiceHolder) Value() Choice {
	return h.value
}

// Get Returns a copy of ChoiceHolder.
func (h ChoiceHolder) Get() interface{} {
	return h
}

// ChoiceFlag A cli Flag that holds a Choice.
type ChoiceFlag struct {
	Name        string
	Aliases     []string
	Value       *ChoiceHolder
	Decoder     ChoiceDecoder
	EnvVars     []string
	FilePath    string
	Usage       string
	DefaultText string
	Required    bool
	Destination *Choice
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

	if f.Value == nil {
		f.Value = NewChoiceHolder(nil)
	}

	if v, ok := flagFromEnvOrFile(f.EnvVars, f.FilePath); ok {
		v, err := f.Decoder.FromString(v)
		if err != nil {
			return fmt.Errorf("supported values: %s", f.Decoder.Strings())
		}
		f.Value = NewChoiceHolder(v)
		f.HasBeenSet = true
	}

	f.Value.init(f.Decoder)

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Var(newChoiceValue(f.Decoder, f.Value, f.Destination), name, f.Usage)
			continue
		}
		set.Var(f.Value, name, f.Usage)
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
	return f.Value.String()
}

// Choice looks up the value of a local ChoiceFlag.
// Returns nil if not found.
func (c *Context) Choice(name string) Choice {
	v := c.Value(name)
	if h, ok := v.(ChoiceHolder); ok {
		return h.Value()
	}
	return nil
}

type choiceValue struct {
	value   *Choice
	decoder ChoiceDecoder
}

func newChoiceValue(decoder ChoiceDecoder, val Choice, p *Choice) *choiceValue {
	*p = val
	return &choiceValue{
		value:   p,
		decoder: decoder,
	}
}

func (c *choiceValue) Set(s string) error {
	log.Printf("called: %s", s)
	v, err := c.decoder.FromString(s)
	if err != nil {
		err = errParse
	}
	*c.value = v
	return err
}

func (c *choiceValue) Get() interface{} { return *c.value }

func (c *choiceValue) String() string {
	if c.value == nil {
		return ""
	}
	return (*c.value).String()
}
