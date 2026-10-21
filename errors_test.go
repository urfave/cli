package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleExitCoder_nil(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	HandleExitCoder(nil)

	assert.Equal(t, 0, exitCode)
	assert.False(t, called)
}

func TestHandleExitCoder_ExitCoder(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	HandleExitCoder(Exit("galactic perimeter breach", 9))

	assert.Equal(t, 9, exitCode)
	assert.True(t, called)
}

func TestHandleExitCoder_ErrorExitCoder(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	HandleExitCoder(Exit(errors.New("galactic perimeter breach"), 9))

	assert.Equal(t, 9, exitCode)
	assert.True(t, called)
}

func TestHandleExitCoder_MultiErrorWithExitCoder(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	exitErr := Exit("galactic perimeter breach", 9)
	exitErr2 := Exit("last ExitCoder", 11)

	err := newMultiError(errors.New("wowsa"), errors.New("egad"), exitErr, exitErr2)
	HandleExitCoder(err)

	assert.Equal(t, 11, exitCode)
	assert.True(t, called)
}

// captureExit runs HandleExitCoder with OsExiter and ErrWriter replaced and
// reports the exit code it saw and what was written to ErrWriter.
func captureExit(err error) (int, bool, string) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	oldWriter := ErrWriter
	defer func() { ErrWriter = oldWriter }()

	var buf bytes.Buffer
	ErrWriter = &buf
	HandleExitCoder(err)

	return exitCode, called, buf.String()
}

// An ExitCoder keeps its exit code and message when it is wrapped with
// fmt.Errorf("%w"): callers add context to an Action error all the time, and
// [exitError.Unwrap] already lets [errors.Is] and [errors.As] see through the
// ExitCoder, so the other direction should work too.
//
// The test is not parallel because captureExit replaces the package level
// OsExiter and ErrWriter, like the rest of this file.
func TestHandleExitCoder_WrappedExitCoder(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		code, called, out := captureExit(fmt.Errorf("deploy failed: %w", Exit("galactic perimeter breach", 9)))

		assert.True(t, called, "OsExiter was not called for a wrapped ExitCoder")
		assert.Equal(t, 9, code)
		assert.Contains(t, out, "galactic perimeter breach")
	})

	t.Run("in multi error", func(t *testing.T) {
		err := newMultiError(
			errors.New("wowsa"),
			fmt.Errorf("deploy failed: %w", Exit("last ExitCoder", 11)),
		)
		code, called, _ := captureExit(err)

		assert.True(t, called, "OsExiter was not called for a wrapped ExitCoder")
		assert.Equal(t, 11, code)
	})
}

type exitFormatter struct {
	code int
}

func (f *exitFormatter) Format(s fmt.State, verb rune) {
	_, _ = s.Write([]byte("some other special"))
}

func (f *exitFormatter) ExitCode() int {
	return f.code
}

func (f *exitFormatter) Error() string {
	return fmt.Sprintf("my special error code %d", f.code)
}

func TestHandleExitCoder_ErrorFormatter(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	oldWriter := ErrWriter
	var buf bytes.Buffer
	ErrWriter = &buf
	defer func() {
		OsExiter = fakeOsExiter
		ErrWriter = oldWriter
	}()

	exitErr := Exit("galactic perimeter breach", 9)
	exitErr2 := Exit("last ExitCoder", 11)
	exitErr3 := &exitFormatter{code: 12}

	// add some recursion for multi error to fix test coverage
	err := newMultiError(errors.New("wowsa"), errors.New("egad"), exitErr3, newMultiError(exitErr, exitErr2))
	HandleExitCoder(err)

	assert.Equal(t, 11, exitCode)
	assert.True(t, called)
	assert.Contains(t, buf.String(), "some other special")
}

func TestHandleExitCoder_MultiErrorWithoutExitCoder(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	err := newMultiError(errors.New("wowsa"), errors.New("egad"))
	HandleExitCoder(err)

	assert.Equal(t, 1, exitCode)
	assert.True(t, called)
}

// make a stub to not import pkg/errors
type ErrorWithFormat struct {
	error
}

func NewErrorWithFormat(m string) *ErrorWithFormat {
	return &ErrorWithFormat{error: errors.New(m)}
}

func (f *ErrorWithFormat) Format(s fmt.State, verb rune) {
	fmt.Fprintf(s, "This the format: %v", f.error)
}

func TestHandleExitCoder_ErrorWithFormat(t *testing.T) {
	called := false

	OsExiter = func(int) {
		if !called {
			called = true
		}
	}
	ErrWriter = &bytes.Buffer{}

	defer func() {
		OsExiter = fakeOsExiter
		ErrWriter = fakeErrWriter
	}()

	err := Exit(NewErrorWithFormat("I am formatted"), 1)
	HandleExitCoder(err)

	assert.True(t, called)
	assert.Equal(t, ErrWriter.(*bytes.Buffer).String(), "This the format: I am formatted\n")
}

func TestHandleExitCoder_MultiErrorWithFormat(t *testing.T) {
	called := false

	OsExiter = func(int) {
		if !called {
			called = true
		}
	}
	ErrWriter = &bytes.Buffer{}

	defer func() { OsExiter = fakeOsExiter }()

	err := newMultiError(NewErrorWithFormat("err1"), NewErrorWithFormat("err2"))
	HandleExitCoder(err)

	assert.True(t, called)
	assert.Equal(t, ErrWriter.(*bytes.Buffer).String(), "This the format: err1\nThis the format: err2\n")
}

func TestMultiErrorErrorsCopy(t *testing.T) {
	errList := []error{
		errors.New("foo"),
		errors.New("bar"),
		errors.New("baz"),
	}
	me := newMultiError(errList...)
	assert.Equal(t, errList, me.Errors())
}

func TestErrRequiredFlags_Error(t *testing.T) {
	missingFlags := []string{"flag1", "flag2"}
	err := &errRequiredFlags{missingFlags: missingFlags}
	expectedMsg := "Required flags \"flag1, flag2\" not set"
	assert.Equal(t, expectedMsg, err.Error())

	missingFlags = []string{"flag1"}
	err = &errRequiredFlags{missingFlags: missingFlags}
	expectedMsg = "Required flag \"flag1\" not set"
	assert.Equal(t, expectedMsg, err.Error())
}

func TestErrRequiredArguments_Error(t *testing.T) {
	missingArguments := []string{"first", "second"}
	err := &errRequiredArguments{missingArguments: missingArguments}
	expectedMsg := "Required arguments \"first, second\" not set"
	assert.Equal(t, expectedMsg, err.Error())

	missingArguments = []string{"first"}
	err = &errRequiredArguments{missingArguments: missingArguments}
	expectedMsg = "Required argument \"first\" not set"
	assert.Equal(t, expectedMsg, err.Error())
}

func TestHandleExitCoder_ExitCoderEmptyMessage(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	// Capture stderr output
	savedErrWriter := ErrWriter
	var errBuf bytes.Buffer
	ErrWriter = &errBuf
	defer func() { ErrWriter = savedErrWriter }()

	HandleExitCoder(Exit("", 42))

	assert.Equal(t, 42, exitCode)
	assert.True(t, called)
	assert.Empty(t, errBuf.String(), "expected no output to stderr for empty exit message")
}

func TestExit_Unwrap(t *testing.T) {
	sentinel := errors.New("not found")

	// The error given to Exit must stay reachable through the standard
	// wrapping helpers (#1090).
	tests := []struct {
		name    string
		message any
		wantMsg string // message of errors.Unwrap(err)
		wantIs  error  // target for errors.Is, if any
		wantAs  any    // target for errors.As, if any
	}{
		{
			name:    "error",
			message: sentinel,
			wantMsg: "not found",
			wantIs:  sentinel,
		},
		{
			name:    "error chain",
			message: &fs.PathError{Op: "open", Path: "x", Err: fs.ErrNotExist},
			wantMsg: "open x: file does not exist",
			wantIs:  fs.ErrNotExist,
			wantAs:  new(*fs.PathError),
		},
		{
			name:    "non-error message",
			message: "boom",
			wantMsg: "boom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Exit(tt.message, 2)

			assert.EqualError(t, errors.Unwrap(err), tt.wantMsg)
			if tt.wantIs != nil {
				assert.ErrorIs(t, err, tt.wantIs)
			}
			if tt.wantAs != nil {
				assert.ErrorAs(t, err, tt.wantAs)
			}
		})
	}
}

// unwrapMultiError is a MultiError that also implements Unwrap() []error, like
// go.uber.org/multierr's combined errors. HandleExitCoder must treat it as a
// MultiError (last exit code wins) instead of walking into it with errors.As,
// which would find the first ExitCoder.
type unwrapMultiError []error

func (m unwrapMultiError) Error() string   { return errors.Join(m...).Error() }
func (m unwrapMultiError) Errors() []error { return m }
func (m unwrapMultiError) Unwrap() []error { return m }

func TestHandleExitCoder_MultiErrorWithUnwrap(t *testing.T) {
	code, called, _ := captureExit(unwrapMultiError{Exit("first", 2), Exit("second", 3)})

	assert.True(t, called)
	assert.Equal(t, 3, code, "the last ExitCoder in a MultiError must win")
}

// A nested MultiError that is the message of an ExitCoder is still an ExitCoder:
// its code must not be replaced by the code of the errors inside it.
func TestHandleExitCoder_NestedExitCoderKeepsCode(t *testing.T) {
	err := newMultiError(
		errors.New("x"),
		Exit(newMultiError(errors.New("a"), errors.New("b")), 5),
	)
	code, called, _ := captureExit(err)

	assert.True(t, called)
	assert.Equal(t, 5, code)
}

// Inside a MultiError a wrapped MultiError is printed as the wrapper message,
// not unwrapped into the bare messages of its elements.
func TestHandleExitCoder_MultiErrorKeepsWrapperMessage(t *testing.T) {
	err := newMultiError(fmt.Errorf("after: %w", newMultiError(errors.New("a"), errors.New("b"))))
	_, called, out := captureExit(err)

	assert.True(t, called)
	assert.Contains(t, out, "after: a")
}
