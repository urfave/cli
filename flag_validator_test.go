package cli

import (
	"fmt"
	"testing"
)

func TestIntValidatorFunc(t *testing.T) {
	funcTests := []struct {
		value     int
		validator IntValidator
		errStr    string
	}{
		{1, IntLT(2), ""},
		{2, IntGTE(2), ""},
		{2, IntLTE(2), ""},
		{10, IntGT(3), ""},
		{10, IntGT(11), fmt.Sprintf(lteFmt, 10, 11)},
		{12, IntGT(12), fmt.Sprintf(lteFmt, 12, 12)},
		{2, IntLT(2), fmt.Sprintf(gteFmt, 2, 2)},
		{10, IntLT(7), fmt.Sprintf(gteFmt, 10, 7)},
		{4, IntGTE(5), fmt.Sprintf(ltFmt, 4, 5)},
		{7, IntLTE(5), fmt.Sprintf(gtFmt, 7, 5)},
	}
	for _, test := range funcTests {
		err := test.validator.ValidateInt(test.value)
		if err == nil {
			if test.errStr != "" {
				t.Errorf("Expected error but got none for %v", test)
			}
			continue
		}
		if test.errStr != err.Error() {
			t.Errorf("Expected error %v got %v", test.errStr, err.Error())
		}
	}
}

func TestIntSliceValidatorFunc(t *testing.T) {
	funcTests := []struct {
		value     []int
		validator IntSliceValidator
		errStr    string
	}{
		{[]int{1}, IntSliceLenLT(2), ""},
		{[]int{1, 2}, IntSliceLenGTE(2), ""},
		{[]int{1}, IntSliceLenLTE(2), ""},
		{[]int{1, 2, 3, 4}, IntSliceLenGT(3), ""},
		{[]int{1, 2, 3}, IntSliceLenGT(11), fmt.Sprintf(sliceLenFmt+lteFmt, 3, 11)},
		{[]int{1, 2, 3}, IntSliceLenGT(3), fmt.Sprintf(sliceLenFmt+lteFmt, 3, 3)},
		{[]int{1, 2}, IntSliceLenLT(2), fmt.Sprintf(sliceLenFmt+gteFmt, 2, 2)},
		{[]int{1, 2, 3, 4, 5}, IntSliceLenLT(4), fmt.Sprintf(sliceLenFmt+gteFmt, 5, 4)},
		{[]int{1, 2}, IntSliceLenGTE(3), fmt.Sprintf(sliceLenFmt+ltFmt, 2, 3)},
		{[]int{1, 2, 3, 4}, IntSliceLenLTE(3), fmt.Sprintf(sliceLenFmt+gtFmt, 4, 3)},
	}
	for _, test := range funcTests {
		err := test.validator.ValidateIntSlice(test.value)
		if err == nil {
			if test.errStr != "" {
				t.Errorf("Expected error but got none for %v", test)
			}
			continue
		}
		if test.errStr != err.Error() {
			t.Errorf("Expected error %v got %v", test.errStr, err.Error())
		}
	}
}
