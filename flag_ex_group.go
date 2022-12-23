package cli

import "fmt"

// FlagExGroup defines a mutually exclusive flag group
// Multiple option paths can be provided out of which
// only one can be defined on cmdline
// So for example
// [ --foo | [ --bar something --darth somethingelse ] ]
type FlagExGroup struct {
	// Flag list
	Flags [][]Flag

	// whether this group is required
	Required bool
}

func (grp FlagExGroup) check(ctx *Context) error {
	oneSet := false

	for _, grpf := range grp.Flags {
		for _, f := range grpf {
			for _, name := range f.Names() {
				if ctx.IsSet(name) {
					if oneSet {
						return fmt.Errorf("muetx set")
					}
					oneSet = true
					break
				}
			}
			if oneSet {
				break
			}
		}
	}

	if !oneSet && grp.Required {
		return fmt.Errorf("missing required flags")
	}
	return nil
}
