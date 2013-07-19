package cli

import "fmt"
import "flag"

type Flag interface {
  fmt.Stringer
  Apply(*flag.FlagSet)
}

type BoolFlag struct {
  Name string
  Usage string
}

type StringFlag struct {
  Name string
  Value string
  Usage string
}

func (f StringFlag) String() string {
  return fmt.Sprintf("--%v 'string'\t%v", f.Name, f.Usage)
}

func (f StringFlag) Apply(set *flag.FlagSet) {
  set.String(f.Name, f.Value, f.Usage)
}

func (f BoolFlag) String() string {
  return fmt.Sprintf("--%v\t%v", f.Name, f.Usage)
}

func (f BoolFlag) Apply(set *flag.FlagSet) {
  set.Bool(f.Name, false, f.Usage)
}

func flagSet(flags []Flag) *flag.FlagSet {
  set := flag.NewFlagSet(Name, flag.ExitOnError)
  for _, f := range flags {
    f.Apply(set)
  }
  return set
}
