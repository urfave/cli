package cli

import (
	"errors"
	"strings"
)

// Kind identifies the class of an [Error]. Kinds are stable: callers match on
// them, catalogues key messages by them, and personalities map them to exit
// codes.
type Kind uint16

const (
	Internal Kind = iota
	UnknownFlag
	MissingValue
	InvalidValue
	RequiredFlag
	RequiredArg
	FlagConflict
	FlagRequires
	UnknownCommand
	MissingCommand
	TooManyArgs
	Help
	Version
)

var kindNames = [...]string{
	Internal:       "internal",
	UnknownFlag:    "unknown_flag",
	MissingValue:   "missing_value",
	InvalidValue:   "invalid_value",
	RequiredFlag:   "required_flag",
	RequiredArg:    "required_arg",
	FlagConflict:   "flag_conflict",
	FlagRequires:   "flag_requires",
	UnknownCommand: "unknown_command",
	MissingCommand: "missing_command",
	TooManyArgs:    "too_many_args",
	Help:           "help",
	Version:        "version",
}

// String returns the stable snake_case name of the kind.
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "unknown"
}

// Key returns the catalogue key for the kind's message.
func (k Kind) Key() string { return "error." + k.String() }

// Usage reports whether the kind describes a mistake in how the program was
// invoked, as opposed to a failure while running it.
func (k Kind) Usage() bool { return k >= UnknownFlag && k <= TooManyArgs }

// Error is the error type returned for every failure the library itself
// detects. It carries structured fields instead of a formatted message, so the
// message can be translated and a program can inspect it without parsing text.
type Error struct {
	Kind    Kind
	Command string   // name of the command the error happened in
	Flag    string   // canonical flag name, without dashes
	Arg     string   // argument name
	Value   string   // the offending input
	Choices []string // valid values, suggestions, or related flags
	Count   int      // quantity used to pick a plural form
	Err     error    // underlying cause, if any
}

// Sentinels for use with [errors.Is]. Only the Kind is compared.
var (
	ErrUnknownFlag    = &Error{Kind: UnknownFlag}
	ErrMissingValue   = &Error{Kind: MissingValue}
	ErrInvalidValue   = &Error{Kind: InvalidValue}
	ErrRequiredFlag   = &Error{Kind: RequiredFlag}
	ErrRequiredArg    = &Error{Kind: RequiredArg}
	ErrFlagConflict   = &Error{Kind: FlagConflict}
	ErrFlagRequires   = &Error{Kind: FlagRequires}
	ErrUnknownCommand = &Error{Kind: UnknownCommand}
	ErrMissingCommand = &Error{Kind: MissingCommand}
	ErrTooManyArgs    = &Error{Kind: TooManyArgs}
	ErrHelp           = &Error{Kind: Help}
	ErrVersion        = &Error{Kind: Version}
)

// Error renders the message in English. Use [Command.Localize] to render it
// in the command's language.
func (e *Error) Error() string { return English.Message(e.Kind.Key(), e.Args()) }

func (e *Error) Unwrap() error { return e.Err }

// Is reports whether target is an *Error of the same Kind.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.Kind == e.Kind
}

// Args returns the values a catalogue message can refer to.
func (e *Error) Args() Args {
	a := Args{
		Command: e.Command,
		Flag:    e.Flag,
		Arg:     e.Arg,
		Value:   e.Value,
		Choices: strings.Join(e.Choices, ", "),
		Count:   e.Count,
	}
	if e.Err != nil {
		a.Cause = e.Err.Error()
	}
	return a
}

// ExitCoder is implemented by errors that choose their own exit code.
type ExitCoder interface {
	error
	ExitCode() int
}

type exitError struct {
	err  error
	code int
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }
func (e *exitError) ExitCode() int { return e.code }

// Exit wraps err so that the program exits with code. The wrapped error stays
// visible to [errors.Is] and [errors.As].
func Exit(err error, code int) error {
	if err == nil {
		return nil
	}
	return &exitError{err: err, code: code}
}

// Split returns the errors joined in err by [errors.Join], or err alone.
func Split(err error) []error {
	if err == nil {
		return nil
	}
	if j, ok := err.(interface{ Unwrap() []error }); ok {
		return j.Unwrap()
	}
	return []error{err}
}

// IsUsage reports whether err, or any error it wraps, is a usage error.
func IsUsage(err error) bool {
	for _, e := range Split(err) {
		var ce *Error
		if errors.As(e, &ce) && ce.Kind.Usage() {
			return true
		}
	}
	return false
}
