package cli

import (
	"flag"
	"time"
)

type extFlag struct {
	f        *flag.Flag
	category string
	// whether this run set the flag, and its value before that
	hasBeenSet bool
	orig       string
}

func (e *extFlag) PreParse() error {
	// A Go flag belongs to another package, so only undo what the previous
	// run set, and keep a value the program set itself, e.g. with flag.Set.
	// Restoring sets the old text again, so a value that rejects it, such
	// as klog's -log_backtrace_at (":0"), keeps the value that run set.
	if e.hasBeenSet {
		e.hasBeenSet = false
		_ = e.f.Value.Set(e.orig)
	}

	return nil
}

func (e *extFlag) PostParse() error {
	return nil
}

func (e *extFlag) Set(_ string, val string) error {
	orig := e.f.Value.String()
	if err := e.f.Value.Set(val); err != nil {
		return err
	}
	if !e.hasBeenSet {
		e.hasBeenSet = true
		e.orig = orig
	}
	return nil
}

func (e *extFlag) Get() any {
	return e.f.Value.(flag.Getter).Get()
}

func (e *extFlag) Names() []string {
	return []string{e.f.Name}
}

func (e *extFlag) IsSet() bool {
	return e.hasBeenSet
}

func (e *extFlag) String() string {
	return FlagStringer(e)
}

func (e *extFlag) IsVisible() bool {
	return true
}

func (e *extFlag) TakesValue() bool {
	return !e.IsBoolFlag()
}

func (e *extFlag) IsBoolFlag() bool {
	bf, ok := e.f.Value.(boolFlag)
	return ok && bf.IsBoolFlag()
}

func (e *extFlag) IsDefaultVisible() bool {
	return e.f.DefValue != ""
}

func (e *extFlag) TypeName() string {
	name, _ := flag.UnquoteUsage(e.f)
	return name
}

func (e *extFlag) IsLocal() bool {
	return false
}

func (e *extFlag) GetCategory() string {
	return e.category
}

func (e *extFlag) SetCategory(c string) {
	e.category = c
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

func (e *extFlag) SchemaType() string {
	switch e.Get().(type) {
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case string:
		return "string"
	case time.Duration:
		return "duration"
	case time.Time:
		return "date-time"
	default:
		return ""
	}
}

func (e *extFlag) SchemaItemsType() string {
	return ""
}
