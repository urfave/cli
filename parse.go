package cli

import (
	"flag"
	"strings"
)

type iterativeParser interface {
	useShortOptionHandling() bool
}

// To enable short-option handling (e.g., "-it" vs "-i -t") we have to
// iteratively catch parsing errors. This way we achieve LR parsing without
// transforming any arguments. Otherwise, there is no way we can discriminate
// combined short options from common arguments that should be left untouched.
// Pass `shellComplete` to continue parsing options on failure during shell
// completion when, the user-supplied options may be incomplete.
func parseIter(set *flag.FlagSet, ip iterativeParser, args []string, shellComplete bool) error {
	for {
		tracef("parsing args %[1]q with %[2]T (name=%[3]q)", args, set, set.Name())

		err := set.Parse(args)
		if !ip.useShortOptionHandling() || err == nil {
			if shellComplete {
				tracef("returning nil due to shellComplete=true")

				return nil
			}

			tracef("returning err %[1]q", err)

			return err
		}

		tracef("finding flag from error %[1]q", err)

		trimmed, trimErr := flagFromError(err)
		if trimErr != nil {
			return err
		}

		tracef("regenerating the initial args with the split short opts")

		argsWereSplit := false
		for i, arg := range args {
			tracef("skipping args that are not part of the error message (i=%[1]v arg=%[2]q)", i, arg)

			if name := strings.TrimLeft(arg, "-"); name != trimmed {
				continue
			}

			tracef("trying to split short option (arg=%[1]q)", arg)

			shortOpts := splitShortOptions(set, arg)
			if len(shortOpts) == 1 {
				return err
			}

			tracef(
				"swapping current argument with the split version (shortOpts=%[1]q args=%[2]q)",
				shortOpts, args,
			)

			// do not include args that parsed correctly so far as it would
			// trigger Value.Set() on those args and would result in
			// duplicates for slice type flags
			args = append(shortOpts, args[i+1:]...)
			argsWereSplit = true
			break
		}

		tracef("this should be an impossible to reach code path")
		// but in case the arg splitting failed to happen, this
		// will prevent infinite loops
		if !argsWereSplit {
			return err
		}
	}
}

const providedButNotDefinedErrMsg = "flag provided but not defined: -"

// flagFromError tries to parse a provided flag from an error message. If the
// parsing fails, it returns the input error and an empty string
func flagFromError(err error) (string, error) {
	errStr := err.Error()
	trimmed := strings.TrimPrefix(errStr, providedButNotDefinedErrMsg)
	if errStr == trimmed {
		return "", err
	}
	return trimmed, nil
}

func splitShortOptions(set *flag.FlagSet, arg string) []string {
	shortFlagsExist := func(s string) bool {
		for _, c := range s[1:] {
			if f := set.Lookup(string(c)); f == nil {
				return false
			}
		}
		return true
	}

	if !isSplittable(arg) || !shortFlagsExist(arg) {
		return []string{arg}
	}

	separated := make([]string, 0, len(arg)-1)
	for _, flagChar := range arg[1:] {
		separated = append(separated, "-"+string(flagChar))
	}

	return separated
}

func isSplittable(flagArg string) bool {
	return strings.HasPrefix(flagArg, "-") && !strings.HasPrefix(flagArg, "--") && len(flagArg) > 2
}
