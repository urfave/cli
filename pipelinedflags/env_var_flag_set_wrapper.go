package altinputsource

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

type EnvVarFlagSetWrapper struct {
	wrappedFsw   cli.FlagSetWrapper
	envVarsMap   map[string]string
	allowedFlags map[string]bool
}

func NewEnvVarFlagSetWrapper(fsw cli.FlagSetWrapper, flags []cli.Flag) cli.FlagSetWrapper {
	envVarsMap := map[string]string{}
	allowedFlags := map[string]bool{}
	for _, f := range flags {
		fise, implementsType := f.(FlagInputSourceExtension)
		if implementsType {

			envVarsMap[f.GetName()] = fise.getEnvVar()
			allowedFlags[f.GetName()] = true
		}
	}

	return &EnvVarFlagSetWrapper{wrappedFsw: fsw, envVarsMap: envVarsMap, allowedFlags: allowedFlags}
}

// Determines if the flag was actually set
func (fsm *EnvVarFlagSetWrapper) HasFlag(name string) bool {
	return fsm.wrappedFsw.HasFlag(name)
}

// Determines if the flag was actually set
func (fsm *EnvVarFlagSetWrapper) IsSet(name string) bool {
	return fsm.wrappedFsw.IsSet(name)
}

// Returns the number of flags set
func (fsm *EnvVarFlagSetWrapper) NumFlags() int {
	return fsm.wrappedFsw.NumFlags()
}

// Returns the command line arguments associated with the context.
func (fsm *EnvVarFlagSetWrapper) Args() cli.Args {
	return fsm.wrappedFsw.Args()
}

func (fsm *EnvVarFlagSetWrapper) Int(name string) int {

	value := fsm.wrappedFsw.Int(name)
	if value == 0 && fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					envValInt, err := strconv.ParseInt(envVal, 0, 64)
					if err == nil {
						return int(envValInt)
					}
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) Duration(name string) time.Duration {
	value := fsm.wrappedFsw.Duration(name)
	if value == 0 && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					envValDuration, err := time.ParseDuration(envVal)
					if err == nil {
						return envValDuration
					}
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) Float64(name string) float64 {
	value := fsm.wrappedFsw.Float64(name)
	if value == 0 && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					envValFloat, err := strconv.ParseFloat(envVal, 10)
					if err == nil {
						return float64(envValFloat)
					}
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) String(name string) string {
	value := fsm.wrappedFsw.String(name)
	if value == "" && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					return envVal
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) StringSlice(name string) []string {
	value := fsm.wrappedFsw.StringSlice(name)
	if value == nil && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					newVal := &cli.StringSlice{}
					for _, s := range strings.Split(envVal, ",") {
						s = strings.TrimSpace(s)
						newVal.Set(s)
					}
					return newVal.Value()
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) IntSlice(name string) []int {
	value := fsm.wrappedFsw.IntSlice(name)
	if value == nil && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					newVal := &cli.IntSlice{}
					for _, s := range strings.Split(envVal, ",") {
						s = strings.TrimSpace(s)
						err := newVal.Set(s)
						if err != nil {
							fmt.Fprintf(os.Stderr, err.Error())
						}
					}

					return newVal.Value()
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) Generic(name string) interface{} {
	value := fsm.wrappedFsw.Generic(name)
	if value == nil && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		var val cli.Generic
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					val.Set(envVal)
					break
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) Bool(name string) bool {
	value := fsm.wrappedFsw.Bool(name)
	if !value && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					envValBool, err := strconv.ParseBool(envVal)
					if err == nil {
						return envValBool
					}
				}
			}
		}
	}

	return value
}

func (fsm *EnvVarFlagSetWrapper) BoolT(name string) bool {
	value := fsm.wrappedFsw.BoolT(name)
	if value && !fsm.allowedFlags[name] {
		envVars := fsm.envVarsMap[name]
		if envVars != "" {
			for _, envVar := range strings.Split(envVars, ",") {
				envVar = strings.TrimSpace(envVar)
				if envVal := os.Getenv(envVar); envVal != "" {
					envValBool, err := strconv.ParseBool(envVal)
					if err == nil {
						return envValBool
					}
				}
			}
		}
	}

	return value
}
