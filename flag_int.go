package cli

import (
	"flag"
	"strconv"
)

// -- int Value
type intValue int

<<<<<<< HEAD
// GetDefaultText returns the default text for this flag
func (f *IntFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return fmt.Sprintf("%d", f.defaultValue)
}

// Apply populates the flag given the flag set and environment
func (f *IntFlag) Apply(set *flag.FlagSet) error {
	// set default value so that environment wont be able to overwrite it
	f.defaultValue = f.Value

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			valInt, err := strconv.ParseInt(val, f.Base, 64)

			if err != nil {
				return fmt.Errorf("could not parse %q as int value from %s for flag %s: %s", val, source, f.Name, err)
			}

			f.Value = int(valInt)
			f.HasBeenSet = true
		}
=======
func (i intValue) Create(val int, p *int) flag.Value {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return err
>>>>>>> Rebase
	}
	*i = intValue(v)
	return err
}

func (i *intValue) Get() any { return int(*i) }

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

type IntFlag = flagImpl[int, intValue]
