package cli

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

var _ Value = (*externalValue)(nil)

var (
	_ Flag                        = (*extFlag)(nil)
	_ ActionableFlag              = (*extFlag)(nil)
	_ CategorizableFlag           = (*extFlag)(nil)
	_ DocGenerationFlag           = (*extFlag)(nil)
	_ DocGenerationMultiValueFlag = (*extFlag)(nil)
	_ LocalFlag                   = (*extFlag)(nil)
	_ RequiredFlag                = (*extFlag)(nil)
	_ VisibleFlag                 = (*extFlag)(nil)
)

// -- Value Value
type externalValue struct {
	e *extFlag
}

// Below functions are to satisfy the flag.Value interface

func (ev *externalValue) Set(s string) error {
	if ev != nil && ev.e.f != nil {
		return ev.e.f.Value.Set(s)
	}
	return nil
}

func (ev *externalValue) Get() any {
	if ev != nil && ev.e.f != nil {
		return ev.e.f.Value.(flag.Getter).Get()
	}
	return nil
}

func (ev *externalValue) String() string {
	if ev != nil && ev.e.f != nil {
		return ev.e.String()
	}
	return ""
}

func (ev *externalValue) IsBoolFlag() bool {
	if ev == nil || ev.e.f == nil {
		return false
	}
	bf, ok := ev.e.f.Value.(boolFlag)
	return ok && bf.IsBoolFlag()
}

type extFlag struct {
	f        *flag.Flag
	category string
}

func (e *extFlag) PreParse() error {
	if e.f.DefValue != "" {
		// suppress errors for write-only external flags that always return nil
		if err := e.Set("", e.f.DefValue); err != nil && e.f.Value.(flag.Getter).Get() != nil {
			// wrap error with some context for the user
			return fmt.Errorf("external flag --%s default %q: %w", e.f.Name, e.f.DefValue, err)
		}
	}

	return nil
}

func (e *extFlag) PostParse() error {
	return nil
}

func (e *extFlag) Set(_ string, val string) error {
	return e.f.Value.Set(val)
}

func (e *extFlag) Get() any {
	return e.f.Value.(flag.Getter).Get()
}

func (e *extFlag) Names() []string {
	return []string{e.f.Name}
}

// IsBoolFlag returns whether the flag doesn't need to accept args
func (e *extFlag) IsBoolFlag() bool {
	if e == nil || e.f == nil {
		return false
	}
	return (&externalValue{e}).IsBoolFlag()
}

// IsDefaultVisible returns true if the flag is not hidden, otherwise false
func (e *extFlag) IsDefaultVisible() bool {
	return true
}

// IsLocal returns false if flag needs to be persistent across subcommands
func (e *extFlag) IsLocal() bool {
	return false
}

// IsMultiValueFlag returns true if the value type T can take multiple
// values from cmd line. This is true for slice and map type flags
func (e *extFlag) IsMultiValueFlag() bool {
	if e == nil || e.f == nil {
		return false
	}
	// TBD how to specify
	if reflect.TypeOf(e.f.Value) == nil {
		return false
	}
	kind := reflect.TypeOf(e.f.Value).Kind()
	return kind == reflect.Slice || kind == reflect.Map
}

// IsRequired returns whether or not the flag is required
func (e *extFlag) IsRequired() bool {
	return false
}

func (e *extFlag) IsSet() bool {
	return false
}

func (e *extFlag) String() string {
	return FlagStringer(e)
}

func (e *extFlag) IsVisible() bool {
	return true
}

func (e *extFlag) TakesValue() bool {
	return false
}

func (e *extFlag) GetUsage() string {
	return e.f.Usage
}

func (e *extFlag) GetValue() string {
	return e.f.Value.String()
}

func (e *extFlag) GetDefaultText() string {
	return e.f.DefValue
}

func (e *extFlag) GetEnvVars() []string {
	return nil
}

// RunAction executes flag action if set
func (e *extFlag) RunAction(ctx context.Context, cmd *Command) error {
	return nil
}

// TypeName returns the type of the flag.
func (e *extFlag) TypeName() string {
	if e == nil || e.f == nil {
		return ""
	}
	ty := reflect.TypeOf(e.f.Value)
	if ty == nil {
		return ""
	}
	// convert the typename to generic type
	convertToGenericType := func(name string) string {
		prefixMap := map[string]string{
			"float": "float",
			"int":   "int",
			"uint":  "uint",
		}
		for prefix, genericType := range prefixMap {
			if strings.HasPrefix(name, prefix) {
				return genericType
			}
		}
		return strings.ToLower(name)
	}

	switch ty.Kind() {
	// if it is a Slice, then return the slice's inner type. Will nested slices be used in the future?
	case reflect.Slice:
		elemType := ty.Elem()
		return convertToGenericType(elemType.Name())
	// if it is a Map, then return the map's key and value types.
	case reflect.Map:
		keyType := ty.Key()
		valueType := ty.Elem()
		return fmt.Sprintf("%s=%s", convertToGenericType(keyType.Name()), convertToGenericType(valueType.Name()))
	default:
		return convertToGenericType(ty.Name())
	}
}

// GetCategory returns the category of the flag
func (e *extFlag) GetCategory() string {
	if e == nil {
		return ""
	}
	return e.category
}

func (e *extFlag) SetCategory(c string) {
	if e != nil {
		e.category = c
	}
}
