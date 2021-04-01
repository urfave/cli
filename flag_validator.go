package cli

import (
	"fmt"
)

const (
	ltFmt         = "%v is less than %v"
	lteFmt        = "%v is less than or equal to %v"
	gtFmt         = "%v is greater than %v"
	gteFmt        = "%v is greater than or equal to %v"
	notInRangeFmt = "%v is not in between %v and %v"
	sliceLenFmt   = "slice len "
)

// Validator is an interface that allows us to mark flags as supporting validation
type Validator interface {

	// Validate retrieves the value from the given context and return an error if value
	// doesnt match its validation rules
	Validate(*Context) error
}

// IntValidator is an interface that allows us to validate int types
type IntValidator interface {

	// ValidateInt validates the given input and returns an error if validation fails
	ValidateInt(int) error
}

// IntSliceValidator is an interface that allows us to validate int slices
type IntSliceValidator interface {

	// ValidateIntSlice validates the given input and returns an error if validation fails
	ValidateIntSlice([]int) error
}

// IntValidatorFunc is a function for int validation
type IntValidatorFunc func(int) error

// ValidateInt function implemenation allows IntValidatorFunc to implement the
// IntValidator interface
func (f IntValidatorFunc) ValidateInt(i int) error {
	return f(i)
}

// ValidateIntSlice function implemenation allows IntValidatorFunc to implement the
// IntValidatorSlice interface
// TODO: Have a way to combine errors into one error string
func (f IntValidatorFunc) ValidateIntSlice(i []int) error {
	var err []error
	for _, item := range i {
		if ierr := f(item); ierr != nil {
			err = append(err, ierr)
		}
	}
	if len(err) == 0 {
		return nil
	}
	return newMultiError(err...)
}

// IntSliceValidatorFunc is a function for int slice validation
type IntSliceValidatorFunc func([]int) error

// ValidateIntSlice function implementation allows IntSliceValidatorFunc to implement the
// IntSliceValidator interface
func (f IntSliceValidatorFunc) ValidateIntSlice(i []int) error {
	return f(i)
}

// CombinedIntSliceValidators allows an array of int slice validators to be
// chained for validation
type CombinedIntSliceValidators []IntSliceValidator

// ValidateIntSlice function implementation allows IntSliceValidatorFunc to implement the
// IntSliceValidator interface
func (cisv CombinedIntSliceValidators) ValidateIntSlice(i []int) error {
	var err []error
	for _, f := range cisv {
		if ierr := f.ValidateIntSlice(i); ierr != nil {
			err = append(err, ierr)
		}
	}
	if len(err) == 0 {
		return nil
	}
	return newMultiError(err...)
}

func IntLT(n int) IntValidatorFunc {
	return func(i int) error {
		if i < n {
			return nil
		}
		return fmt.Errorf(gteFmt, i, n)
	}
}

func IntLTE(n int) IntValidatorFunc {
	return func(i int) error {
		if i <= n {
			return nil
		}
		return fmt.Errorf(gtFmt, i, n)
	}
}

func IntGT(n int) IntValidatorFunc {
	return func(i int) error {
		if i > n {
			return nil
		}
		return fmt.Errorf(lteFmt, i, n)
	}
}

func IntGTE(n int) IntValidatorFunc {
	return func(i int) error {
		if i >= n {
			return nil
		}
		return fmt.Errorf(ltFmt, i, n)
	}
}

func IntInRange(a, b int) IntValidatorFunc {
	return func(i int) error {
		if i >= a && i <= b {
			return nil
		}
		return fmt.Errorf(notInRangeFmt, i, a, b)
	}
}

func IntSliceLenLT(n int) IntSliceValidator {
	return IntSliceValidatorFunc(func(i []int) error {
		if len(i) >= n {
			return fmt.Errorf(sliceLenFmt+gteFmt, len(i), n)
		}
		return nil
	})
}

func IntSliceLenLTE(n int) IntSliceValidator {
	return IntSliceValidatorFunc(func(i []int) error {
		if len(i) > n {
			return fmt.Errorf(sliceLenFmt+gtFmt, len(i), n)
		}
		return nil
	})
}

func IntSliceLenGT(n int) IntSliceValidator {
	return IntSliceValidatorFunc(func(i []int) error {
		if len(i) <= n {
			return fmt.Errorf(sliceLenFmt+lteFmt, len(i), n)
		}
		return nil
	})
}

func IntSliceLenGTE(n int) IntSliceValidator {
	return IntSliceValidatorFunc(func(i []int) error {
		if len(i) < n {
			return fmt.Errorf(sliceLenFmt+ltFmt, len(i), n)
		}
		return nil
	})
}

func IntSliceLenInRange(a, b int) IntSliceValidator {
	return IntSliceValidatorFunc(func(i []int) error {
		if len(i) >= a && len(i) <= b {
			return nil
		}
		return fmt.Errorf(sliceLenFmt+notInRangeFmt, len(i), a, b)
	})
}
