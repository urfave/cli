package cli

import (
	"flag"
	"fmt"
)

type BoolWithInverseFlag interface {
	fmt.Stringer

	// Apply Flag settings to the given flag set
	Apply(*flag.FlagSet) error

	// All possible names for this flag
	Names() []string

	// Whether the flag has been set or not
	IsSet() bool

	// Will return the original flag and the inverse flag
	Flags() []Flag

	// Will return the value as it relates to the boolFlag provided
	Value() bool
}

type boolWithInverse struct {
	fmt.Stringer
	positiveFlag *BoolFlag
	negativeFlag *BoolFlag

	action func(*Context, bool) error

	posDest *bool
	negDest *bool

	posCount *int
	negCount *int
}

func (s *boolWithInverse) Flags() []Flag {
	return []Flag{s.positiveFlag, s.negativeFlag}
}

func (s *boolWithInverse) IsSet() bool {
	return *s.negDest || *s.posDest
}

func (s *boolWithInverse) Value() bool {
	return *s.posDest
}

func (s *boolWithInverse) RunAction(ctx *Context) error {
	if *s.negDest && *s.posDest {
		return fmt.Errorf("cannot set both flags `--%s` and `--%s`", s.positiveFlag.Name, s.negativeFlag.Name)
	}

	if *s.negDest {
		err := ctx.Set(s.positiveFlag.Name, "false")
		if err != nil {
			return err
		}
	} else if *s.posDest {
		err := ctx.Set(s.negativeFlag.Name, "false")
		if err != nil {
			return err
		}
	}

	if s.positiveFlag.Action != nil {
		return s.positiveFlag.Action(ctx, *s.posDest)
	}

	return nil
}

/*
NewBoolWithInverse creates a new BoolFlag that has an inverse flag

consider a bool flag `--env`, there is no way to set it to false
this function allows you to set `--env` or `--no-env` and in the command action
it can be determined that BoolWithInverseFlag.IsSet()
*/
func NewBoolWithInverse(flag BoolFlag) BoolWithInverseFlag {
	special := &boolWithInverse{
		negDest:  new(bool),
		negCount: new(int),
	}

	if flag.Destination != nil {
		special.posDest = flag.Destination
	} else {
		special.posDest = new(bool)
	}

	if flag.Count != nil {
		special.posCount = flag.Count
	} else {
		special.posCount = new(int)
	}

	special.positiveFlag = &flag
	special.positiveFlag.Destination = special.posDest
	special.positiveFlag.Count = special.posCount

	// Append `no-` to each alias
	var inverseAliases []string
	if len(flag.Aliases) > 0 {
		inverseAliases = make([]string, len(flag.Aliases))
		for idx, alias := range flag.Aliases {
			inverseAliases[idx] = "no-" + alias
		}
	}

	special.negativeFlag = &BoolFlag{
		Name:        "no-" + flag.Name,
		Category:    flag.Category,
		DefaultText: flag.DefaultText,
		FilePath:    flag.FilePath,
		Usage:       flag.Usage,
		Required:    flag.Required,
		Hidden:      flag.Hidden,
		HasBeenSet:  flag.HasBeenSet,
		Value:       flag.Value,
		Aliases:     inverseAliases,

		Destination: special.negDest,
		Count:       special.negCount,
	}

	if len(flag.EnvVars) > 0 {
		// TODO we need to append to the action to reverse the value of the env vars
		special.negativeFlag.EnvVars = append([]string{}, flag.EnvVars...)
	}

	return special
}

func (s *boolWithInverse) Apply(set *flag.FlagSet) error {
	if err := s.positiveFlag.Apply(set); err != nil {
		return err
	}

	if err := s.negativeFlag.Apply(set); err != nil {
		return err
	}

	return nil
}

func (s *boolWithInverse) Names() []string {
	if *s.negDest {
		return s.negativeFlag.Names()
	}

	if *s.posDest {
		return s.positiveFlag.Names()
	}

	return append(s.negativeFlag.Names(), s.positiveFlag.Names()...)
}

func (s *boolWithInverse) String() string {
	return fmt.Sprintf("%s || %s", s.positiveFlag.String(), s.negativeFlag.String())
}
