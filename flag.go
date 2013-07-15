package cli

import "fmt"

type Flag interface {
  fmt.Stringer
}

type BoolFlag struct {
  Name string
  Value bool
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

func (f BoolFlag) String() string {
  return fmt.Sprintf("--%v\t%v", f.Name, f.Usage)
}
