// Package cli provides a minimal framework for creating and organizing command line
// Go applications. cli is designed to be easy to understand and write, the most simple
// cli application can be written as follows:
//   func main() {
//     cli.NewApp().Run(os.Args)
//   }
//
// Of course this application does not do much, so let's make this an actual application:
//   func main() {
//     app := cli.NewApp()
//     app.Name = "greet"
//     app.Usage = "say a greeting"
//     app.Action = func(c *cli.Context) {
//       println("Greetings")
//     }
//
//     app.Run(os.Args)
//   }
package cli

import (
	"fmt"
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
