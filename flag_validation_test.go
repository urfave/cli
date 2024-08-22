package cli

import (
	"fmt"
	"testing"

	itesting "github.com/urfave/cli/v3/internal/testing"
)

func TestFlagDefaultValidation(t *testing.T) {
	cmd := &Command{
		Name: "foo",
		Flags: []Flag{
			&IntFlag{
				Name:  "if",
				Value: 2, // this value should fail validation
				Validator: func(i int64) error {
					if (i >= 3 && i <= 10) || (i >= 20 && i <= 24) {
						return nil
					}
					return fmt.Errorf("Value %d not in range [3,10] or [20,24]", i)
				},
			},
		},
	}

	// Default value of flag is 2 which should fail validation
	err := cmd.Run(buildTestContext(t), []string{"foo", "--if", "5"})
	itesting.RequireError(t, err)
}

func TestFlagValidation(t *testing.T) {
	testCases := []struct {
		name        string
		arg         string
		errExpected bool
	}{
		{
			name:        "first range less than min",
			arg:         "2",
			errExpected: true,
		},
		{
			name: "first range min",
			arg:  "3",
		},
		{
			name: "first range mid",
			arg:  "7",
		},
		{
			name: "first range max",
			arg:  "10",
		},
		{
			name:        "first range greater than max",
			arg:         "15",
			errExpected: true,
		},
		{
			name:        "second range less than min",
			arg:         "19",
			errExpected: true,
		},
		{
			name: "second range min",
			arg:  "20",
		},
		{
			name: "second range mid",
			arg:  "21",
		},
		{
			name: "second range max",
			arg:  "24",
		},
		{
			name:        "second range greater than max",
			arg:         "27",
			errExpected: true,
		},
	}

	for _, testCase := range testCases {
		cmd := &Command{
			Name: "foo",
			Flags: []Flag{
				&IntFlag{
					Name:  "it",
					Value: 5, // note that this value should pass validation
					Validator: func(i int64) error {
						if (i >= 3 && i <= 10) || (i >= 20 && i <= 24) {
							return nil
						}
						return fmt.Errorf("Value %d not in range [3,10]U[20,24]", i)
					},
				},
			},
		}

		err := cmd.Run(buildTestContext(t), []string{"foo", "--it", testCase.arg})
		if !testCase.errExpected {
			itesting.RequireNoError(t, err)
		} else {
			itesting.RequireError(t, err)
		}
	}
}
