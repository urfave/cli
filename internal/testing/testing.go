package testing

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"
)

func NoError(t *testing.T, err error, msgAndArgs ...interface{}) bool {
	if err != nil {
		t.Helper()
		return fail(t, fmt.Sprintf("Received unexpected error:\n%+v", err), msgAndArgs...)
	}

	return true
}

func RequireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if NoError(t, err, msgAndArgs...) {
		return
	}

	t.FailNow()
}

func Error(t *testing.T, err error, msgAndArgs ...interface{}) bool {
	if err == nil {
		t.Helper()
		return fail(t, fmt.Sprintf("Received unexpected error:\n%+v", err), msgAndArgs...)
	}

	return true
}

func RequireError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()

	if Error(t, err, msgAndArgs...) {
		return
	}

	t.FailNow()
}

func ErrorContains(t *testing.T, theError error, contains string, msgAndArgs ...interface{}) bool {
	t.Helper()

	if !Error(t, theError, msgAndArgs...) {
		return false
	}

	actual := theError.Error()
	if !strings.Contains(actual, contains) {
		return fail(t, fmt.Sprintf("Error %#v does not contain %#v", actual, contains), msgAndArgs...)
	}

	return true
}

func RequireErrorContains(t *testing.T, theError error, contains string, msgAndArgs ...interface{}) {
	t.Helper()

	if ErrorContains(t, theError, contains, msgAndArgs...) {
		return
	}

	t.FailNow()
}

func ErrorIs(t *testing.T, err, target error, msgAndArgs ...interface{}) bool {
	t.Helper()

	if errors.Is(err, target) {
		return true
	}

	var expectedText string
	if target != nil {
		expectedText = target.Error()
	}

	chain := buildErrorChainString(err)

	return fail(t, fmt.Sprintf("Target error should be in err chain:\n"+
		"expected: %q\n"+
		"in chain: %s", expectedText, chain,
	), msgAndArgs...)
}

func EqualError(t *testing.T, theError error, errString string, msgAndArgs ...interface{}) bool {
	t.Helper()

	if !Error(t, theError, msgAndArgs...) {
		return false
	}
	expected := errString
	actual := theError.Error()
	// don't need to use deep equals here, we know they are both strings
	if expected != actual {
		return fail(t, fmt.Sprintf("Error message not equal:\n"+
			"expected: %q\n"+
			"actual  : %q", expected, actual), msgAndArgs...)
	}
	return true
}

func RequireEqualError(t *testing.T, theError error, errString string, msgAndArgs ...interface{}) {
	t.Helper()

	if EqualError(t, theError, errString, msgAndArgs...) {
		return
	}

	t.FailNow()
}

func Equal(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if err := validateEqualArgs(expected, actual); err != nil {
		return fail(t, fmt.Sprintf("Invalid operation: %#v == %#v (%s)",
			expected, actual, err), msgAndArgs...)
	}

	if !ObjectsAreEqual(expected, actual) {
		expected, actual = formatUnequalValues(expected, actual)
		return fail(t, fmt.Sprintf("Not equal: \n"+
			"expected: %s\n"+
			"actual  : %s", expected, actual), msgAndArgs...)
	}

	return true
}

func RequireEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	if Equal(t, expected, actual, msgAndArgs...) {
		return
	}

	t.FailNow()
}

func Equalf(t *testing.T, expected interface{}, actual interface{}, msg string, args ...interface{}) bool {
	t.Helper()

	return Equal(t, expected, actual, append([]interface{}{msg}, args...)...)
}

func RequireEqualf(t *testing.T, expected interface{}, actual interface{}, msg string, args ...interface{}) {
	t.Helper()

	if Equalf(t, expected, actual, msg, args...) {
		return
	}

	t.FailNow()
}

func Contains(t *testing.T, s, contains interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	ok, found := containsElement(s, contains)
	if !ok {
		return fail(t, fmt.Sprintf("%#v could not be applied builtin len()", s), msgAndArgs...)
	}
	if !found {
		return fail(t, fmt.Sprintf("%#v does not contain %#v", s, contains), msgAndArgs...)
	}

	return true
}

func NotContains(t *testing.T, s, contains interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	ok, found := containsElement(s, contains)
	if !ok {
		return fail(t, fmt.Sprintf("%#v could not be applied builtin len()", s), msgAndArgs...)
	}
	if found {
		return fail(t, fmt.Sprintf("%#v should not contain %#v", s, contains), msgAndArgs...)
	}

	return true
}

func True(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	if !value {
		t.Helper()
		return fail(t, "Should be true", msgAndArgs...)
	}

	return true
}

func False(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	if value {
		t.Helper()
		return fail(t, "Should be false", msgAndArgs...)
	}

	return true
}

func Empty(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	pass := isEmpty(object)
	if !pass {
		t.Helper()
		return fail(t, fmt.Sprintf("Should be empty, but was %v", object), msgAndArgs...)
	}

	return pass
}

func NotEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	pass := !isEmpty(object)
	if !pass {
		t.Helper()
		return fail(t, fmt.Sprintf("Should NOT be empty, but was %v", object), msgAndArgs...)
	}

	return pass
}

func Nil(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	if isNil(object) {
		return true
	}
	t.Helper()
	return fail(t, fmt.Sprintf("Expected nil, but got: %#v", object), msgAndArgs...)
}

func NotNil(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	if !isNil(object) {
		return true
	}
	t.Helper()
	return fail(t, "Expected value not to be nil.", msgAndArgs...)
}

func Zero(t *testing.T, i interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if i != nil && !reflect.DeepEqual(i, reflect.Zero(reflect.TypeOf(i)).Interface()) {
		return fail(t, fmt.Sprintf("Should be zero, but was %v", i), msgAndArgs...)
	}

	return true
}

func NotZero(t *testing.T, i interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if i == nil || reflect.DeepEqual(i, reflect.Zero(reflect.TypeOf(i)).Interface()) {
		return fail(t, fmt.Sprintf("Should not be zero, but was %v", i), msgAndArgs...)
	}

	return true
}

func Len(t *testing.T, object interface{}, length int, msgAndArgs ...interface{}) bool {
	t.Helper()

	ok, l := getLen(object)
	if !ok {
		return fail(t, fmt.Sprintf("\"%s\" could not be applied builtin len()", object), msgAndArgs...)
	}

	if l != length {
		return fail(t, fmt.Sprintf("\"%s\" should have %d item(s), but has %d", object, length, l), msgAndArgs...)
	}
	return true
}

func RequireLen(t *testing.T, object interface{}, length int, msgAndArgs ...interface{}) {
	t.Helper()

	if Len(t, object, length, msgAndArgs...) {
		return
	}

	t.FailNow()
}

func JSONEq(t *testing.T, expected string, actual string, msgAndArgs ...interface{}) bool {
	t.Helper()

	var expectedJSONAsInterface, actualJSONAsInterface interface{}

	if err := json.Unmarshal([]byte(expected), &expectedJSONAsInterface); err != nil {
		return fail(t, fmt.Sprintf("Expected value ('%s') is not valid json.\nJSON parsing error: '%s'", expected, err.Error()), msgAndArgs...)
	}

	if err := json.Unmarshal([]byte(actual), &actualJSONAsInterface); err != nil {
		return fail(t, fmt.Sprintf("Input ('%s') needs to be valid json.\nJSON parsing error: '%s'", actual, err.Error()), msgAndArgs...)
	}

	return Equal(t, expectedJSONAsInterface, actualJSONAsInterface, msgAndArgs...)
}

// fail reports a failure through
func fail(t *testing.T, failureMessage string, msgAndArgs ...interface{}) bool {
	t.Helper()
	content := []labeledContent{
		{"Test", t.Name()},
		{"Error Trace", strings.Join(callerInfo(), "\n\t\t\t")},
		{"Error", failureMessage},
	}

	message := messageFromMsgAndArgs(msgAndArgs...)
	if len(message) > 0 {
		content = append(content, labeledContent{"Messages", message})
	}

	t.Errorf("\n%s", ""+labeledOutput(content...))

	return false
}

type labeledContent struct {
	label   string
	content string
}

// labeledOutput returns a string consisting of the provided labeledContent. Each labeled output is appended in the following manner:
//
//	\t{{label}}:{{align_spaces}}\t{{content}}\n
//
// The initial carriage return is required to undo/erase any padding added by testing.T.Errorf. The "\t{{label}}:" is for the label.
// If a label is shorter than the longest label provided, padding spaces are added to make all the labels match in length. Once this
// alignment is achieved, "\t{{content}}\n" is added for the output.
//
// If the content of the labeledOutput contains line breaks, the subsequent lines are aligned so that they start at the same location as the first line.
func labeledOutput(content ...labeledContent) string {
	longestLabel := 0
	for _, v := range content {
		if len(v.label) > longestLabel {
			longestLabel = len(v.label)
		}
	}
	var output string
	for _, v := range content {
		output += "\t" + v.label + ":" + strings.Repeat(" ", longestLabel-len(v.label)) + "\t" + indentMessageLines(v.content, longestLabel) + "\n"
	}
	return output
}

/* callerInfo is necessary because the assert functions use the testing object
internally, causing it to print the file:line of the assert method, rather than where
the problem actually occurred in calling code.*/

// callerInfo returns an array of strings containing the file and line number
// of each stack frame leading from the current test to the assert call that
// failed.
func callerInfo() []string {

	var pc uintptr
	var ok bool
	var file string
	var line int
	var name string

	callers := []string{}
	for i := 0; ; i++ {
		pc, file, line, ok = runtime.Caller(i)
		if !ok {
			// The breaks below failed to terminate the loop, and we ran off the
			// end of the call stack.
			break
		}

		// This is a huge edge case, but it will panic if this is the case, see #180
		if file == "<autogenerated>" {
			break
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		name = f.Name()

		// testing.tRunner is the standard library function that calls
		// tests. Subtests are called directly by tRunner, without going through
		// the Test/Benchmark/Example function that contains the t.Run calls, so
		// with subtests we should break when we hit tRunner, without adding it
		// to the list of callers.
		if name == "testing.tRunner" {
			break
		}

		parts := strings.Split(file, "/")
		if len(parts) > 1 {
			filename := parts[len(parts)-1]
			dir := parts[len(parts)-2]
			if (dir != "assert" && dir != "mock" && dir != "require") || filename == "mock_test.go" {
				callers = append(callers, fmt.Sprintf("%s:%d", file, line))
			}
		}

		// Drop the package
		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}

	return callers
}

// Stolen from the `go test` tool.
// isTest tells whether name looks like a test (or benchmark, according to prefix).
// It is a Test (say) if there is a character after Test that is not a lower-case letter.
// We don't want TesticularCancer.
func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) { // "Test" is ok
		return true
	}
	r, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(r)
}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}

// Aligns the provided message so that all lines after the first line start at the same location as the first line.
// Assumes that the first line starts at the correct location (after carriage return, tab, label, spacer and tab).
// The longestLabelLen parameter specifies the length of the longest label in the output (required becaues this is the
// basis on which the alignment occurs).
func indentMessageLines(message string, longestLabelLen int) string {
	outBuf := new(bytes.Buffer)

	for i, scanner := 0, bufio.NewScanner(strings.NewReader(message)); scanner.Scan(); i++ {
		// no need to align first line because it starts at the correct location (after the label)
		if i != 0 {
			// append alignLen+1 spaces to align with "{{longestLabel}}:" before adding tab
			outBuf.WriteString("\n\t" + strings.Repeat(" ", longestLabelLen+1) + "\t")
		}
		outBuf.WriteString(scanner.Text())
	}

	return outBuf.String()
}

func buildErrorChainString(err error) string {
	if err == nil {
		return ""
	}

	e := errors.Unwrap(err)
	chain := fmt.Sprintf("%q", err.Error())
	for e != nil {
		chain += fmt.Sprintf("\n\t%q", e.Error())
		e = errors.Unwrap(e)
	}
	return chain
}

// validateEqualArgs checks whether provided arguments can be safely used in the
// Equal/NotEqual functions.
func validateEqualArgs(expected, actual interface{}) error {
	if expected == nil && actual == nil {
		return nil
	}

	if isFunction(expected) || isFunction(actual) {
		return errors.New("cannot take func type as argument")
	}
	return nil
}

func isFunction(arg interface{}) bool {
	if arg == nil {
		return false
	}
	return reflect.TypeOf(arg).Kind() == reflect.Func
}

// formatUnequalValues takes two values of arbitrary types and returns string
// representations appropriate to be presented to the user.
//
// If the values are not of like type, the returned strings will be prefixed
// with the type name, and the value will be enclosed in parenthesis similar
// to a type conversion in the Go grammar.
func formatUnequalValues(expected, actual interface{}) (e string, a string) {
	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		return fmt.Sprintf("%T(%s)", expected, truncatingFormat(expected)),
			fmt.Sprintf("%T(%s)", actual, truncatingFormat(actual))
	}
	switch expected.(type) {
	case time.Duration:
		return fmt.Sprintf("%v", expected), fmt.Sprintf("%v", actual)
	}
	return truncatingFormat(expected), truncatingFormat(actual)
}

// truncatingFormat formats the data and truncates it if it's too long.
//
// This helps keep formatted error messages lines from exceeding the
// bufio.MaxScanTokenSize max line length that the go testing framework imposes.
func truncatingFormat(data interface{}) string {
	value := fmt.Sprintf("%#v", data)
	max := bufio.MaxScanTokenSize - 100 // Give us some space the type info too if needed.
	if len(value) > max {
		value = value[0:max] + "<... truncated>"
	}
	return value
}

// ObjectsAreEqual determines if two objects are considered equal.
//
// This function does no assertion of any kind.
func ObjectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

func containsElement(list interface{}, element interface{}) (ok, found bool) {

	listValue := reflect.ValueOf(list)
	listType := reflect.TypeOf(list)
	if listType == nil {
		return false, false
	}
	listKind := listType.Kind()
	defer func() {
		if e := recover(); e != nil {
			ok = false
			found = false
		}
	}()

	if listKind == reflect.String {
		elementValue := reflect.ValueOf(element)
		return true, strings.Contains(listValue.String(), elementValue.String())
	}

	if listKind == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if ObjectsAreEqual(mapKeys[i].Interface(), element) {
				return true, true
			}
		}
		return true, false
	}

	for i := 0; i < listValue.Len(); i++ {
		if ObjectsAreEqual(listValue.Index(i).Interface(), element) {
			return true, true
		}
	}
	return true, false
}

// isEmpty gets whether the specified object is considered empty or not.
func isEmpty(object interface{}) bool {

	// get nil case out of the way
	if object == nil {
		return true
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
	// pointers are empty if nil or if the value they point to is empty
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
	// for all other types, compare against the zero value
	// array types are empty when they match their zero-initialized state
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	isNilableKind := containsKind(
		[]reflect.Kind{
			reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map,
			reflect.Ptr, reflect.Slice, reflect.UnsafePointer},
		kind)

	if isNilableKind && value.IsNil() {
		return true
	}

	return false
}

func containsKind(kinds []reflect.Kind, kind reflect.Kind) bool {
	for i := 0; i < len(kinds); i++ {
		if kind == kinds[i] {
			return true
		}
	}

	return false
}

// getLen try to get length of object.
// return (false, 0) if impossible.
func getLen(x interface{}) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}
