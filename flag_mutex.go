package cli

// MutuallyExclusiveFlags defines a mutually exclusive flag group
// Multiple option paths can be provided out of which
// only one can be defined on cmdline
// So for example
// [ --foo | [ --bar something --darth somethingelse ] ]
type MutuallyExclusiveFlags struct {
	// Flag list
	Flags [][]Flag

	// whether this group is required
	Required bool

	// Category to apply to all flags within group
	Category string
}

func (grp MutuallyExclusiveFlags) check(cmd *Command) error {
	oneSet := false
	e := &mutuallyExclusiveGroup{}

	for _, grpf := range grp.Flags {
		for _, f := range grpf {
			for _, name := range f.Names() {
				if cmd.IsSet(name) {
					if oneSet {
						e.flag2Name = name
						return e
					}
					e.flag1Name = name
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
		return &mutuallyExclusiveGroupRequiredFlag{flags: &grp}
	}
	return nil
}

func (grp MutuallyExclusiveFlags) propagateCategory() {
	for _, grpf := range grp.Flags {
		for _, f := range grpf {
			if cf, ok := f.(CategorizableFlag); ok {
				cf.SetCategory(grp.Category)
			}
		}
	}
}
