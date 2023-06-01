package cli

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
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

	expect(t, exitCode, 0)
	expect(t, called, false)
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

	expect(t, exitCode, 9)
	expect(t, called, true)
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

	expect(t, exitCode, 9)
	expect(t, called, true)
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

	expect(t, exitCode, 11)
	expect(t, called, true)
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

	expect(t, exitCode, 1)
	expect(t, called, true)
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

	expect(t, called, true)
	expect(t, ErrWriter.(*bytes.Buffer).String(), "This the format: I am formatted\n")
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

	expect(t, called, true)
	expect(t, ErrWriter.(*bytes.Buffer).String(), "This the format: err1\nThis the format: err2\n")
}

func TestMultiErrorErrorsCopy(t *testing.T) {
	errList := []error{
		errors.New("foo"),
		errors.New("bar"),
		errors.New("baz"),
	}
	me := newMultiError(errList...)
	expect(t, errList, me.Errors())
}
