package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// OsExiter is the function used when the app exits. If not set defaults to os.Exit.
var OsExiter = os.Exit

// ErrWriter is used to write errors to the user. This can be anything
// implementing the io.Writer interface and defaults to os.Stderr.
var ErrWriter io.Writer = os.Stderr

// MultiError is an error that wraps multiple errors.
type MultiError interface {
	error
	// Errors returns a copy of the errors slice
	Errors() []error
}

// NewMultiError creates a new MultiError. Pass in one or more errors.
func newMultiError(err ...error) MultiError {
	ret := multiError(err)
	return &ret
}

type multiError []error

// Error implements the error interface.
func (m *multiError) Error() string {
	errs := make([]string, len(*m))
	for i, err := range *m {
		errs[i] = err.Error()
	}

	return strings.Join(errs, "\n")
}

// Errors returns a copy of the errors slice
func (m *multiError) Errors() []error {
	errs := make([]error, len(*m))
	for _, err := range *m {
		errs = append(errs, err)
	}
	return errs
}

// ExitCoder is the interface checked by `App` and `Command` for a custom exit
// code
type ExitCoder interface {
	error
	ExitCode() int
}

type exitError struct {
	exitCode int
	message  string
}

// Exit wraps a message and exit code into an ExitCoder suitable for handling by
// HandleExitCoder
func Exit(message string, exitCode int) ExitCoder {
	return &exitError{
		exitCode: exitCode,
		message:  message,
	}
}

func (ee *exitError) Error() string {
	return ee.message
}

func (ee *exitError) ExitCode() int {
	return ee.exitCode
}

// HandleExitCoder checks if the error fulfills the ExitCoder interface, and if
// so prints the error to stderr (if it is non-empty) and calls OsExiter with the
// given exit code.  If the given error is a MultiError, then this func is
// called on all members of the Errors slice.
func HandleExitCoder(err error) {
	if err == nil {
		return
	}

	if exitErr, ok := err.(ExitCoder); ok {
		if err.Error() != "" {
			fmt.Fprintln(ErrWriter, err)
		}
		OsExiter(exitErr.ExitCode())
		return
	}

	if multiErr, ok := err.(MultiError); ok {
		for _, merr := range multiErr.Errors() {
			HandleExitCoder(merr)
		}
		return
	}

	if err.Error() != "" {
		fmt.Fprintln(ErrWriter, err)
	}
	OsExiter(1)
}
