package cli

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

// ConditionOrError ir a helper function to make writing
// validation functions much easier
func ConditionOrError(cond bool, err error) error {
	if cond {
		return nil
	}
	return err
}

// ValidationChain allows one to chain a sequence of validation
// functions to construct a single validation function.
func ValidationChain[T any](fns ...func(T) error) func(T) error {
	return func(v T) error {
		for _, fn := range fns {
			if err := fn(v); err != nil {
				return err
			}
		}
		return nil
	}
}

// Min means that the value to be checked needs to be atleast(and including)
// the checked value
func Min[T constraints.Ordered](c T) func(T) error {
	return func(v T) error {
		return ConditionOrError(v >= c, fmt.Errorf("%v is not less than %v", v, c))
	}
}

// Max means that the value to be checked needs to be atmost(and including)
// the checked value
func Max[T constraints.Ordered](c T) func(T) error {
	return func(v T) error {
		return ConditionOrError(v <= c, fmt.Errorf("%v is not greater than %v", v, c))
	}
}

// Max means that the value to be checked needs to be atmost(and including)
// the checked value
func RangeInclusive[T constraints.Ordered](a, b T) func(T) error {
	return ValidationChain[T](Min[T](a), Max[T](b))
}
