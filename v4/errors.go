package cli

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Class groups kinds by how a program should treat them.
type Class uint8

const (
	// Failure means something went wrong while running. It exits 1.
	Failure Class = iota
	// Usage means the program was invoked wrongly. It exits with the
	// personality's usage code.
	Usage
)

// Help and version requests are not errors. Run renders them and returns nil.

// Kind identifies what went wrong in an [Error]. Callers match on kinds,
// catalogues key messages by them, and personalities map them to exit codes.
//
// A kind carries its own name and class, so each one is defined in a single
// place. Create kinds with [NewKind] at package level. The zero Kind is
// [Internal].
type Kind struct {
	name  string
	class Class
}

var (
	kinds    = []Kind{Internal}
	kindSeen = map[string]bool{"internal": true}
)

// NewKind defines a kind. name is the stable snake_case name used in
// catalogue keys and JSON output. Applications can define their own kinds for
// errors they want translated or matched. NewKind panics if name is taken, so
// call it at package level where a clash shows up at startup.
func NewKind(name string, class Class) Kind {
	if name == "" || kindSeen[name] {
		panic("cli: duplicate or empty error kind " + strconv.Quote(name))
	}
	k := Kind{name: name, class: class}
	kindSeen[name] = true
	kinds = append(kinds, k)
	return k
}

// Kinds returns every defined kind, in definition order.
func Kinds() []Kind { return slices.Clone(kinds) }

// The kinds the library itself returns.
var (
	Internal Kind // an error that fits no other kind

	UnknownFlag    = NewKind("unknown_flag", Usage)
	MissingValue   = NewKind("missing_value", Usage)
	InvalidValue   = NewKind("invalid_value", Usage)
	RequiredFlag   = NewKind("required_flag", Usage)
	RequiredArg    = NewKind("required_arg", Usage)
	FlagConflict   = NewKind("flag_conflict", Usage)
	FlagRequires   = NewKind("flag_requires", Usage)
	UnknownCommand = NewKind("unknown_command", Usage)
	MissingCommand = NewKind("missing_command", Usage)
	TooManyArgs    = NewKind("too_many_args", Usage)
)

// String returns the kind's stable name.
func (k Kind) String() string {
	if k.name == "" {
		return "internal"
	}
	return k.name
}

// Key returns the catalogue key for the kind's message.
func (k Kind) Key() string { return "error." + k.String() }

// Class reports how a program should treat the kind.
func (k Kind) Class() Class { return k.class }

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

// Exit returns an error that makes the program exit with code. message is
// usually a string or an error; an error stays visible to [errors.Is] and
// [errors.As]. The signature matches v3, so existing calls compile unchanged.
func Exit(message any, code int) ExitCoder {
	err, ok := message.(error)
	if !ok {
		err = errors.New(fmt.Sprint(message))
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
		if errors.As(e, &ce) && ce.Kind.Class() == Usage {
			return true
		}
	}
	return false
}
