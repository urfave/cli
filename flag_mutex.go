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

func (grp MutuallyExclusiveFlags) check(_ *Command) error {
	e := &mutuallyExclusiveGroup{}

	// First, find the index of the group were the flag was set
	// (if it exists.)
	var i int
	var oneSet bool

flagGroupLoop:
	for ; i < len(grp.Flags); i++ {
		group := grp.Flags[i]

		// For each flag inside this group, check if it's set.
		for _, flg := range group {
			if flg.IsSet() {
				e.flag1Name = flg.Names()[0]
				oneSet = true
				break flagGroupLoop
			}
		}
	}

	// Next, continue from the flag group just after the one we
	// stopped at above, to see if another flag is set. If so,
	// return an error.
	i++
	for ; i < len(grp.Flags); i++ {
		group := grp.Flags[i]

		for _, flg := range group {
			if flg.IsSet() {
				e.flag2Name = flg.Names()[0]
				return e
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
