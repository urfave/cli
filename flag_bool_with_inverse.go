package cli

import (
	"flag"
	"fmt"
	"strings"
)

var (
	DefaultInverseBoolPrefix = "no-"
)

type BoolWithInverseFlag struct {
	// The BoolFlag which the positive and negative flags are generated from
	*BoolFlag

	// The prefix used to indicate a negative value
	// Default: `env` becomes `no-env`
	InversePrefix string

	positiveFlag *BoolFlag
	negativeFlag *BoolFlag

	// pointers obtained from the embedded bool flag
	posDest  *bool
	posCount *int

	negDest *bool
}

func (s *BoolWithInverseFlag) Flags() []Flag {
	return []Flag{s.positiveFlag, s.negativeFlag}
}

func (s *BoolWithInverseFlag) IsSet() bool {
	return (*s.posCount > 0) || (s.positiveFlag.IsSet() || s.negativeFlag.IsSet())
}

func (s *BoolWithInverseFlag) Value() bool {
	return *s.posDest
}

func (s *BoolWithInverseFlag) RunAction(ctx *Context) error {
	if *s.negDest && *s.posDest {
		return fmt.Errorf("cannot set both flags `--%s` and `--%s`", s.positiveFlag.Name, s.negativeFlag.Name)
	}

	if *s.negDest {
		err := ctx.Set(s.positiveFlag.Name, "false")
		if err != nil {
			return err
		}
	}

	if s.BoolFlag.Action != nil {
		return s.BoolFlag.Action(ctx, s.Value())
	}

	return nil
}

/*
initialize creates a new BoolFlag that has an inverse flag

consider a bool flag `--env`, there is no way to set it to false
this function allows you to set `--env` or `--no-env` and in the command action
it can be determined that BoolWithInverseFlag.IsSet()
*/
func (parent *BoolWithInverseFlag) initialize() {
	child := parent.BoolFlag

	parent.negDest = new(bool)
	if child.Destination != nil {
		parent.posDest = child.Destination
	} else {
		parent.posDest = new(bool)
	}

	if child.Config.Count != nil {
		parent.posCount = child.Config.Count
	} else {
		parent.posCount = new(int)
	}

	parent.positiveFlag = child
	parent.positiveFlag.Destination = parent.posDest
	parent.positiveFlag.Config.Count = parent.posCount

	parent.negativeFlag = &BoolFlag{
		Category:    child.Category,
		DefaultText: child.DefaultText,
		FilePaths:   append([]string{}, child.FilePaths...),
		Usage:       child.Usage,
		Required:    child.Required,
		Hidden:      child.Hidden,
		Persistent:  child.Persistent,
		Value:       child.Value,
		Destination: parent.negDest,
		TakesFile:   child.TakesFile,
		OnlyOnce:    child.OnlyOnce,
		hasBeenSet:  child.hasBeenSet,
		applied:     child.applied,
		creator:     boolValue{},
		value:       child.value,
	}

	// Set inverse names ex: --env => --no-env
	parent.negativeFlag.Name = parent.inverseName()
	parent.negativeFlag.Aliases = parent.inverseAliases()

	if len(child.EnvVars) > 0 {
		parent.negativeFlag.EnvVars = make([]string, len(child.EnvVars))
		for idx, envVar := range child.EnvVars {
			parent.negativeFlag.EnvVars[idx] = strings.ToUpper(parent.InversePrefix) + envVar
		}
	}

	return
}

func (parent *BoolWithInverseFlag) inverseName() string {
	if parent.InversePrefix == "" {
		parent.InversePrefix = DefaultInverseBoolPrefix
	}

	return parent.InversePrefix + parent.BoolFlag.Name
}

func (parent *BoolWithInverseFlag) inverseAliases() (aliases []string) {
	if len(parent.BoolFlag.Aliases) > 0 {
		aliases = make([]string, len(parent.BoolFlag.Aliases))
		for idx, alias := range parent.BoolFlag.Aliases {
			aliases[idx] = parent.InversePrefix + alias
		}
	}

	return
}

func (s *BoolWithInverseFlag) Apply(set *flag.FlagSet) error {
	if s.positiveFlag == nil {
		s.initialize()
	}

	if err := s.positiveFlag.Apply(set); err != nil {
		return err
	}

	if err := s.negativeFlag.Apply(set); err != nil {
		return err
	}

	return nil
}

func (s *BoolWithInverseFlag) Names() []string {
	// Get Names when flag has not been initialized
	if s.positiveFlag == nil {
		return append(s.BoolFlag.Names(), FlagNames(s.inverseName(), s.inverseAliases())...)
	}

	if *s.negDest {
		return s.negativeFlag.Names()
	}

	if *s.posDest {
		return s.positiveFlag.Names()
	}

	return append(s.negativeFlag.Names(), s.positiveFlag.Names()...)
}

// Example for BoolFlag{Name: "env"}
// --env     (default: false) || --no-env    (default: false)
func (s *BoolWithInverseFlag) String() string {
	if s.positiveFlag == nil {
		return fmt.Sprintf("%s || --%s", s.BoolFlag.String(), s.inverseName())
	}

	return fmt.Sprintf("%s || %s", s.positiveFlag.String(), s.negativeFlag.String())
}
