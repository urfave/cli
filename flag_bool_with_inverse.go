package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"
)

var DefaultInverseBoolPrefix = "no-"

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

func (parent *BoolWithInverseFlag) Flags() []Flag {
	return []Flag{parent.positiveFlag, parent.negativeFlag}
}

func (parent *BoolWithInverseFlag) IsSet() bool {
	return (*parent.posCount > 0) || (parent.positiveFlag.IsSet() || parent.negativeFlag.IsSet())
}

func (parent *BoolWithInverseFlag) Value() bool {
	return *parent.posDest
}

func (parent *BoolWithInverseFlag) RunAction(ctx context.Context, cmd *Command) error {
	if *parent.negDest && *parent.posDest {
		return fmt.Errorf("cannot set both flags `--%s` and `--%s`", parent.positiveFlag.Name, parent.negativeFlag.Name)
	}

	if *parent.negDest {
		_ = cmd.Set(parent.positiveFlag.Name, "false")
	}

	if parent.BoolFlag.Action != nil {
		return parent.BoolFlag.Action(ctx, cmd, parent.Value())
	}

	return nil
}

// Initialize creates a new BoolFlag that has an inverse flag
//
// consider a bool flag `--env`, there is no way to set it to false
// this function allows you to set `--env` or `--no-env` and in the command action
// it can be determined that BoolWithInverseFlag.IsSet().
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
		Sources:     NewValueSourceChain(child.Sources.Chain...),
		Usage:       child.Usage,
		Required:    child.Required,
		Hidden:      child.Hidden,
		Local:       child.Local,
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

	if len(child.Sources.EnvKeys()) > 0 {
		sources := []ValueSource{}

		for _, envVar := range child.GetEnvVars() {
			sources = append(sources, EnvVar(strings.ToUpper(parent.InversePrefix)+envVar))
		}
		parent.negativeFlag.Sources = NewValueSourceChain(sources...)
	}
}

func (parent *BoolWithInverseFlag) inverseName() string {
	return parent.inversePrefix() + parent.BoolFlag.Name
}

func (parent *BoolWithInverseFlag) inversePrefix() string {
	if parent.InversePrefix == "" {
		parent.InversePrefix = DefaultInverseBoolPrefix
	}

	return parent.InversePrefix
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

func (parent *BoolWithInverseFlag) Apply(set *flag.FlagSet) error {
	if parent.positiveFlag == nil {
		parent.initialize()
	}

	if err := parent.positiveFlag.Apply(set); err != nil {
		return err
	}

	if err := parent.negativeFlag.Apply(set); err != nil {
		return err
	}

	return nil
}

func (parent *BoolWithInverseFlag) Names() []string {
	// Get Names when flag has not been initialized
	if parent.positiveFlag == nil {
		return append(parent.BoolFlag.Names(), FlagNames(parent.inverseName(), parent.inverseAliases())...)
	}

	if *parent.negDest {
		return parent.negativeFlag.Names()
	}

	if *parent.posDest {
		return parent.positiveFlag.Names()
	}

	return append(parent.negativeFlag.Names(), parent.positiveFlag.Names()...)
}

// String implements the standard Stringer interface.
//
// Example for BoolFlag{Name: "env"}
// --[no-]env	(default: false)
func (parent *BoolWithInverseFlag) String() string {
	out := FlagStringer(parent)
	i := strings.Index(out, "\t")

	prefix := "--"

	// single character flags are prefixed with `-` instead of `--`
	if len(parent.Name) == 1 {
		prefix = "-"
	}

	return fmt.Sprintf("%s[%s]%s%s", prefix, parent.inversePrefix(), parent.Name, out[i:])
}
