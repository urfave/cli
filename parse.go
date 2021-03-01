package cli

import (
	"flag"
	"fmt"
	"strings"
)

type iterativeParser interface {
	newFlagSet() (*flag.FlagSet, error)
	useShortOptionHandling() bool
	collectUnusedFlags() bool
	setUnusedFlags([]string)
}

// To enable short-option handling (e.g., "-it" vs "-i -t") we have to
// iteratively catch parsing errors. This way we achieve LR parsing without
// transforming any arguments. Otherwise, there is no way we can discriminate
// combined short options from common arguments that should be left untouched.
// Pass `shellComplete` to continue parsing options on failure during shell
// completion when, the user-supplied options may be incomplete.
func parseIter(set *flag.FlagSet, ip iterativeParser, args []string, shellComplete bool) error {
	var unusedArgs []string
	defer func() {
		ip.setUnusedFlags(unusedArgs)
	}()
	for {
		if ip.useShortOptionHandling() && ip.collectUnusedFlags() {
			return fmt.Errorf("can not setup short option handling and unused flag collecting the same time")
		}

		err := set.Parse(args)
		if err == nil {
			return nil
		}
		if !ip.useShortOptionHandling() && !ip.collectUnusedFlags() {
			if shellComplete {
				return nil
			}
			return err
		}

		errStr := err.Error()
		if ip.useShortOptionHandling() {
			trimmed := strings.TrimPrefix(errStr, "flag provided but not defined: -")
			if errStr == trimmed {
				return err
			}

			// regenerate the initial args with the split short opts
			argsWereSplit := false
			for i, arg := range args {
				// skip args that are not part of the error message
				if name := strings.TrimLeft(arg, "-"); name != trimmed {
					continue
				}

				// if we can't split, the error was accurate
				shortOpts := splitShortOptions(set, arg)
				if len(shortOpts) == 1 {
					return err
				}

				// swap current argument with the split version
				args = append(args[:i], append(shortOpts, args[i+1:]...)...)
				argsWereSplit = true
				break
			}

			// This should be an impossible to reach code path, but in case the arg
			// splitting failed to happen, this will prevent infinite loops
			if !argsWereSplit {
				return err
			}

			// Since custom parsing failed, replace the flag set before retrying
			newSet, err := ip.newFlagSet()
			if err != nil {
				return err
			}
			*set = *newSet
			continue
		}

		// regard `--unk  | -unk` prefix as unknown flag
		// --unk (args will be empty)
		// --unk --next-flag ... (args will be --next-flag ...)
		// --unk arg ... -next-flag arg2 ...(args will be -next-flag arg2)
		// --unk1 arg1 ... --unk2 arg2 ... -next-flag arg3 ...(args will be -next-flag arg3)
		if ip.collectUnusedFlags() {
			trimmed := strings.TrimPrefix(errStr, "flag provided but not defined: ")
			if errStr == trimmed {
				return err
			}
			var newArgs []string
			var unknown = false
			for i, arg := range args {
				if arg == trimmed || (arg[0] == '-' && arg == "-" + trimmed) {
					unknown = true
					unusedArgs = append(unusedArgs, arg)
					continue
				}
				if unknown && strings.HasPrefix(arg, "-"){
					newArgs = append(newArgs, args[i:]...)
					break
				}
				if unknown {
					unusedArgs = append(unusedArgs, arg)
				} else {
					newArgs = append(newArgs, arg)
				}
			}
			args = newArgs
		}
	}
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
