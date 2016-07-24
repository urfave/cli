package cli

import (
	"bytes"
	"errors"
	"testing"
)

func TestHandleExitCoder_nil(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		exitCode = rc
		called = true
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
		exitCode = rc
		called = true
	}

	defer func() { OsExiter = fakeOsExiter }()

	HandleExitCoder(NewExitError("galactic perimeter breach", 9))

	expect(t, exitCode, 9)
	expect(t, called, true)
}

func TestHandleExitCoder_MultiErrorWithExitCoder(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		exitCode = rc
		called = true
	}

	defer func() { OsExiter = fakeOsExiter }()

	exitErr := NewExitError("galactic perimeter breach", 9)
	err := NewMultiError(errors.New("wowsa"), errors.New("egad"), exitErr)
	HandleExitCoder(err)

	expect(t, exitCode, 9)
	expect(t, called, true)
}

func TestHandleExitCoder_ErrorWithMessage(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		exitCode = rc
		called = true
	}
	ErrWriter = &bytes.Buffer{}

	defer func() {
		OsExiter = fakeOsExiter
		ErrWriter = fakeErrWriter
	}()

	err := errors.New("gourd havens")
	HandleExitCoder(err)

	expect(t, exitCode, 1)
	expect(t, called, true)
	expect(t, ErrWriter.(*bytes.Buffer).String(), "gourd havens\n")
}

func TestHandleExitCoder_ErrorWithoutMessage(t *testing.T) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		exitCode = rc
		called = true
	}
	ErrWriter = &bytes.Buffer{}

	defer func() {
		OsExiter = fakeOsExiter
		ErrWriter = fakeErrWriter
	}()

	err := errors.New("")
	HandleExitCoder(err)

	expect(t, exitCode, 1)
	expect(t, called, true)
	expect(t, ErrWriter.(*bytes.Buffer).String(), "")
}
