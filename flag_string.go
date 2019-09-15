package cli

import "flag"

// StringFlag is a flag with type string
type StringFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	TakesFile   bool
	Value       string
	DefaultText string
	Destination *string
}

// String returns a readable representation of this value
// (for usage defaults)
func (s *StringFlag) String() string {
	return FlagStringer(s)
}

// Names returns the names of the flag
func (s *StringFlag) Names() []string {
	return flagNames(s)
}

// IsRequired returns whether or not the flag is required
func (s *StringFlag) IsRequired() bool {
	return s.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (s *StringFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (s *StringFlag) GetUsage() string {
	return s.Usage
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (s *StringFlag) GetValue() string {
	return s.Value
}

// Apply populates the flag given the flag set and environment
func (s *StringFlag) Apply(set *flag.FlagSet) error {
	if val, ok := flagFromEnvOrFile(s.EnvVars, s.FilePath); ok {
		s.Value = val
	}

	for _, name := range s.Names() {
		if s.Destination != nil {
			set.StringVar(s.Destination, name, s.Value, s.Usage)
			continue
		}
		set.String(name, s.Value, s.Usage)
	}

	return nil
}

// String looks up the value of a local StringFlag, returns
// "" if not found
func (c *Context) String(name string) string {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupString(name, fs)
	}
	return ""
}

// GlobalString looks up the value of a global StringFlag, returns
// "" if not found
//func (c *Context) GlobalString(name string) string {
//	if fs := lookupGlobalFlagSet(name, c); fs != nil {
//		return lookupPath(name, fs)
//	}
//	return ""
//}

func lookupString(name string, set *flag.FlagSet) string {
	f := set.Lookup(name)
	if f != nil {
		parsed, err := f.Value.String(), error(nil)
		if err != nil {
			return ""
		}
		return parsed
	}
	return ""
}
