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

	// Check for the use of a mutually-exclusive flag, starting at
	// the first group.
	name, i, ok := grp.findSetFlag(0)
	if ok {
		e.flag1Name = name
		i++

		// Check for the use of a flag in a mutually exclusive
		// relationship with the one we just found.
		if name2, _, ok := grp.findSetFlag(i); ok {
			e.flag2Name = name2
			return e
		}
	}

	if !ok && grp.Required {
		return &mutuallyExclusiveGroupRequiredFlag{flags: &grp}
	}

	return nil
}

// findSetFlag is used in [MutuallyExclusiveFlags.check] to find
// whether at least one flag inside a mutually exclusive flag group is
// set. If so, return the flag name, position at which it's set, and
// Boolean true (indicating that a flag was found.) Else, return all
// zero values.
func (grp MutuallyExclusiveFlags) findSetFlag(startIdx int) (string, int, bool) {
	for i := startIdx; i < len(grp.Flags); i++ {
		flags := grp.Flags[i]

		for _, flg := range flags {
			if flg.IsSet() {
				return flg.Names()[0], i, true
			}
		}
	}

	return "", 0, false
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
