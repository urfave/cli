package cli

import (
	"flag"
	"fmt"
	"strings"
)

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
		for index, c := range s[1:] {
			if index == (len(s[1:])-1) && c == '-' {
				break
			}
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
		if flagChar != '-' {
			separated = append(separated, "-"+string(flagChar))
		} else {
			separated = append(separated, "-")
		}
	}

	return separated
}

func isSplittable(flagArg string) bool {
	return strings.HasPrefix(flagArg, "-") && !strings.HasPrefix(flagArg, "--") && len(flagArg) > 2
}

func getFlagNameValue(arg string) (string, string, error) {
	if arg[0] != '-' || len(arg) == 1 {
		return "", "", fmt.Errorf("not a flag")
	}
	numMinus := 1
	if arg[1] == '-' {
		numMinus++
		if len(arg) == 2 {
			return "", "", nil
		}
	}

	arg = arg[numMinus:]
	if index := strings.Index(arg, "="); index == -1 {
		return arg, "", nil
	} else {
		return arg[:index], arg[index:], nil
	}
}
