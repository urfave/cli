package altsrc

import (
	"fmt"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/urfave/cli/v2"
)

// FlagInputSourceExtension is an extension interface of cli.Flag that
// allows a value to be set on the existing parsed flags.
type FlagInputSourceExtension interface {
	cli.Flag
	ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error
}

// ApplyInputSourceValues iterates over all provided flags and
// executes ApplyInputSourceValue on flags implementing the
// FlagInputSourceExtension interface to initialize these flags
// to an alternate input source.
func ApplyInputSourceValues(cCtx *cli.Context, inputSourceContext InputSourceContext, flags []cli.Flag) error {
	for _, f := range flags {
		inputSourceExtendedFlag, isType := f.(FlagInputSourceExtension)
		if isType {
			err := inputSourceExtendedFlag.ApplyInputSourceValue(cCtx, inputSourceContext)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// InitInputSource is used to to setup an InputSourceContext on a cli.Command Before method. It will create a new
// input source based on the func provided. If there is no error it will then apply the new input source to any flags
// that are supported by the input source
func InitInputSource(flags []cli.Flag, createInputSource func() (InputSourceContext, error)) cli.BeforeFunc {
	return func(cCtx *cli.Context) error {
		inputSource, err := createInputSource()
		if err != nil {
			return fmt.Errorf("Unable to create input source: inner error: \n'%v'", err.Error())
		}

		return ApplyInputSourceValues(cCtx, inputSource, flags)
	}
}

// InitInputSourceWithContext is used to to setup an InputSourceContext on a cli.Command Before method. It will create a new
// input source based on the func provided with potentially using existing cli.Context values to initialize itself. If there is
// no error it will then apply the new input source to any flags that are supported by the input source
func InitInputSourceWithContext(flags []cli.Flag, createInputSource func(cCtx *cli.Context) (InputSourceContext, error)) cli.BeforeFunc {
	return func(cCtx *cli.Context) error {
		inputSource, err := createInputSource(cCtx)
		if err != nil {
			return fmt.Errorf("Unable to create input source with context: inner error: \n'%v'", err.Error())
		}

		return ApplyInputSourceValues(cCtx, inputSource, flags)
	}
}

// ApplyInputSourceValue applies a generic value to the flagSet if required
func (f *GenericFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.GenericFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Generic(name)
		if err != nil {
			return err
		}
		if value == nil {
			continue
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, value.String())
		}
	}

	return nil
}

// ApplyInputSourceValue applies a StringSlice value to the flagSet if required
func (f *StringSliceFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.StringSliceFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.StringSlice(name)
		if err != nil {
			return err
		}
		if value == nil {
			continue
		}
		var sliceValue = *(cli.NewStringSlice(value...))
		for _, n := range f.Names() {
			underlyingFlag := f.set.Lookup(n)
			if underlyingFlag == nil {
				continue
			}
			underlyingFlag.Value = &sliceValue
		}
		if f.Destination != nil {
			f.Destination.Set(sliceValue.Serialize())
		}
	}
	return nil
}

// ApplyInputSourceValue applies a IntSlice value if required
func (f *IntSliceFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.IntSliceFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.IntSlice(name)
		if err != nil {
			return err
		}
		if value == nil {
			continue
		}
		var sliceValue = *(cli.NewIntSlice(value...))
		for _, n := range f.Names() {
			underlyingFlag := f.set.Lookup(n)
			if underlyingFlag == nil {
				continue
			}
			underlyingFlag.Value = &sliceValue
		}
		if f.Destination != nil {
			f.Destination.Set(sliceValue.Serialize())
		}
	}
	return nil
}

// ApplyInputSourceValue applies a Int64Slice value if required
func (f *Int64SliceFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.Int64SliceFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Int64Slice(name)
		if err != nil {
			return err
		}
		if value == nil {
			continue
		}
		var sliceValue = *(cli.NewInt64Slice(value...))
		for _, n := range f.Names() {
			underlyingFlag := f.set.Lookup(n)
			if underlyingFlag == nil {
				continue
			}
			underlyingFlag.Value = &sliceValue
		}
		if f.Destination != nil {
			f.Destination.Set(sliceValue.Serialize())
		}
	}
	return nil
}

// ApplyInputSourceValue applies a Float64Slice value if required
func (f *Float64SliceFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.Float64SliceFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Float64Slice(name)
		if err != nil {
			return err
		}
		if value == nil {
			continue
		}
		var sliceValue = *(cli.NewFloat64Slice(value...))
		for _, n := range f.Names() {
			underlyingFlag := f.set.Lookup(n)
			if underlyingFlag == nil {
				continue
			}
			underlyingFlag.Value = &sliceValue
		}
		if f.Destination != nil {
			f.Destination.Set(sliceValue.Serialize())
		}
	}
	return nil
}

// ApplyInputSourceValue applies a Bool value to the flagSet if required
func (f *BoolFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.BoolFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Bool(name)
		if err != nil {
			return err
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, strconv.FormatBool(value))
		}
	}
	return nil
}

// ApplyInputSourceValue applies a String value to the flagSet if required
func (f *StringFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.StringFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.String(name)
		if err != nil {
			return err
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, value)
		}
	}
	return nil
}

// ApplyInputSourceValue applies a Path value to the flagSet if required
func (f *PathFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.PathFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.String(name)
		if err != nil {
			return err
		}
		if value == "" {
			continue
		}
		for _, n := range f.Names() {
			if !filepath.IsAbs(value) && isc.Source() != "" {
				basePathAbs, err := filepath.Abs(isc.Source())
				if err != nil {
					return err
				}
				value = filepath.Join(filepath.Dir(basePathAbs), value)
			}
			_ = f.set.Set(n, value)
		}
	}
	return nil
}

// ApplyInputSourceValue applies a int value to the flagSet if required
func (f *IntFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.IntFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Int(name)
		if err != nil {
			return err
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, strconv.FormatInt(int64(value), 10))
		}
	}
	return nil
}

func (f *Int64Flag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.Int64Flag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Int64(name)
		if err != nil {
			return err
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, strconv.FormatInt(value, 10))
		}
	}
	return nil
}

func (f *UintFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.UintFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Uint(name)
		if err != nil {
			return err
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, strconv.FormatUint(uint64(value), 10))
		}
	}
	return nil
}

func (f *Uint64Flag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.Uint64Flag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Uint64(name)
		if err != nil {
			return err
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, strconv.FormatUint(value, 10))
		}
	}
	return nil
}

// ApplyInputSourceValue applies a Duration value to the flagSet if required
func (f *DurationFlag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.DurationFlag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Duration(name)
		if err != nil {
			return err
		}
		for _, n := range f.Names() {
			_ = f.set.Set(n, value.String())
		}
	}
	return nil
}

// ApplyInputSourceValue applies a Float64 value to the flagSet if required
func (f *Float64Flag) ApplyInputSourceValue(cCtx *cli.Context, isc InputSourceContext) error {
	if f.set == nil || cCtx.IsSet(f.Name) || isEnvVarSet(f.EnvVars) {
		return nil
	}
	for _, name := range f.Float64Flag.Names() {
		if !isc.isSet(name) {
			continue
		}
		value, err := isc.Float64(name)
		if err != nil {
			return err
		}
		floatStr := float64ToString(value)
		for _, n := range f.Names() {
			_ = f.set.Set(n, floatStr)
		}
	}
	return nil
}

func isEnvVarSet(envVars []string) bool {
	for _, envVar := range envVars {
		if _, ok := syscall.Getenv(envVar); ok {
			// TODO: Can't use this for bools as
			// set means that it was true or false based on
			// Bool flag type, should work for other types
			return true
		}
	}

	return false
}

func float64ToString(f float64) string {
	return fmt.Sprintf("%v", f)
}
