package cli

import (
	"fmt"
	"os"
	"strings"
)

type MultiError struct {
	Errors []error
}

func NewMultiError(err ...error) MultiError {
	return MultiError{Errors: err}
}

func (m MultiError) Error() string {
	errs := make([]string, len(m.Errors))
	for i, err := range m.Errors {
		errs[i] = err.Error()
	}

	return strings.Join(errs, "\n")
}

type ExitCoder interface {
	ExitCode() int
}

type ExitError struct {
	exitCode int
	message  string
}

func NewExitError(message string, exitCode int) *ExitError {
	return &ExitError{
		exitCode: exitCode,
		message:  message,
	}
}

func (ee *ExitError) Error() string {
	return ee.message
}

func (ee *ExitError) String() string {
	return fmt.Sprintf("%s exitcode=%v", ee.message, ee.exitCode)
}

func (ee *ExitError) ExitCode() int {
	return ee.exitCode
}

// HandleExitCoder checks if the error fulfills the ExitCoder interface, and if
// so calls os.Exit with the given exit code.
func HandleExitCoder(err error) {
	if exitErr, ok := err.(ExitCoder); ok {
		os.Exit(exitErr.ExitCode())
	}
}
