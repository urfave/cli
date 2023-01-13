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

	action func(*Context, bool) error

	// pointers obtained from the embedded bool flag
	posDest  *bool
	posCount *int

	negDest  *bool
	negCount *int
}

func (s *BoolWithInverseFlag) Flags() []Flag {
	return []Flag{s.positiveFlag, s.negativeFlag}
}

func (s *BoolWithInverseFlag) IsSet() bool {
	return (*s.posCount > 0 || *s.negCount > 0) || (s.positiveFlag.IsSet() || s.negativeFlag.IsSet())
}

func (s *BoolWithInverseFlag) Value() bool {
	return *s.posDest
}

func (s *BoolWithInverseFlag) RunAction(ctx *Context) error {
	if *s.negDest && *s.posDest {
		return fmt.Errorf("cannot set both flags `--%s` and `--%s`", s.positiveFlag.Name, s.negativeFlag.Name)
	}

	if *s.negDest {
		err := ctx.Set(s.negativeFlag.Name, "true")
		if err != nil {
			return err
		}
	} else if *s.posDest {
		err := ctx.Set(s.positiveFlag.Name, "true")
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
	parent.negCount = new(int)
	if child.Destination != nil {
		parent.posDest = child.Destination
	} else {
		parent.posDest = new(bool)
	}

	if child.Count != nil {
		parent.posCount = child.Count
	} else {
		parent.posCount = new(int)
	}

	parent.positiveFlag = child
	parent.positiveFlag.Destination = parent.posDest
	parent.positiveFlag.Count = parent.posCount

	if parent.InversePrefix == "" {
		parent.InversePrefix = DefaultInverseBoolPrefix
	}

	// Append `no-` to each alias
	var inverseAliases []string
	if len(child.Aliases) > 0 {
		inverseAliases = make([]string, len(child.Aliases))
		for idx, alias := range child.Aliases {
			inverseAliases[idx] = parent.InversePrefix + alias
		}
	}

	parent.negativeFlag = &BoolFlag{
		Name:        parent.InversePrefix + child.Name,
		Category:    child.Category,
		DefaultText: child.DefaultText,
		FilePath:    child.FilePath,
		Usage:       child.Usage,
		Required:    child.Required,
		Hidden:      child.Hidden,
		HasBeenSet:  child.HasBeenSet,
		Value:       child.Value,
		Aliases:     inverseAliases,

		Destination: parent.negDest,
		Count:       parent.negCount,
	}

	if len(child.EnvVars) > 0 {
		parent.negativeFlag.EnvVars = make([]string, len(child.EnvVars))
		for idx, envVar := range child.EnvVars {
			parent.negativeFlag.EnvVars[idx] = strings.ToUpper(parent.InversePrefix) + envVar
		}
	}

	return
}

func (s *BoolWithInverseFlag) Apply(set *flag.FlagSet) error {
	s.initialize()

	if err := s.positiveFlag.Apply(set); err != nil {
		return err
	}

	if err := s.negativeFlag.Apply(set); err != nil {
		return err
	}

	return nil
}

func (s *BoolWithInverseFlag) Names() []string {
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
		return s.BoolFlag.String()
	}

	return fmt.Sprintf("%s || %s", s.positiveFlag.String(), s.negativeFlag.String())
}
