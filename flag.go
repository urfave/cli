package cli

import "fmt"
import "flag"

type Flag interface {
	fmt.Stringer
	Apply(*flag.FlagSet)
}

type SliceFlag interface {
	Value() []string
}

func flagSet(name string, flags []Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

type StringSlice []string

func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *StringSlice) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *StringSlice) Value() []string {
	return *i
}

type StringSliceFlag struct {
	Name  string
	Value *StringSlice
	Usage string
}

func (f StringSliceFlag) String() string {
	return fmt.Sprintf("%s%v '%v'\t%v", prefixFor(f.Name), f.Name, "-"+f.Name+" option -"+f.Name+" option", f.Usage)
}

func (f StringSliceFlag) Apply(set *flag.FlagSet) {
	set.Var(f.Value, f.Name, f.Usage)
}

type BoolFlag struct {
	Name  string
	Usage string
}

func (f BoolFlag) String() string {
	return fmt.Sprintf("%s%v\t%v", prefixFor(f.Name), f.Name, f.Usage)
}

func (f BoolFlag) Apply(set *flag.FlagSet) {
	set.Bool(f.Name, false, f.Usage)
}

type StringFlag struct {
	Name  string
	Value string
	Usage string
}

func (f StringFlag) String() string {
	return fmt.Sprintf("%s%v '%v'\t%v", prefixFor(f.Name), f.Name, f.Value, f.Usage)
}

func (f StringFlag) Apply(set *flag.FlagSet) {
	set.String(f.Name, f.Value, f.Usage)
}

type IntFlag struct {
	Name  string
	Value int
	Usage string
}

func (f IntFlag) String() string {
	return fmt.Sprintf("%s%v '%v'\t%v", prefixFor(f.Name), f.Name, f.Value, f.Usage)
}

func (f IntFlag) Apply(set *flag.FlagSet) {
	set.Int(f.Name, f.Value, f.Usage)
}

type helpFlag struct {
	Usage string
}

func (f helpFlag) String() string {
	return fmt.Sprintf("--help, -h\t%v", f.Usage)
}

func (f helpFlag) Apply(set *flag.FlagSet) {
	set.Bool("h", false, f.Usage)
	set.Bool("help", false, f.Usage)
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}
