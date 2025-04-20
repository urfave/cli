package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagDefaultValidation(t *testing.T) {
	cmd := &Command{
		Name: "foo",
		Flags: []Flag{
			&Int64Flag{
				Name:  "if",
				Value: 2, // this value should fail validation
				Validator: func(i int64) error {
					if (i >= 3 && i <= 10) || (i >= 20 && i <= 24) {
						return nil
					}
					return fmt.Errorf("Value %d not in range [3,10] or [20,24]", i)
				},
				ValidateDefaults: true,
			},
		},
	}

	r := require.New(t)

	// this is a simple call to test PreParse failure before
	// parsing has been done
	r.Error(cmd.Set("if", "11"))

	// Default value of flag is 2 which should fail validation
	err := cmd.Run(buildTestContext(t), []string{"foo", "--if", "5"})
	r.Error(err)
}

func TestBoolInverseFlagDefaultValidation(t *testing.T) {
	cmd := &Command{
		Name: "foo",
		Flags: []Flag{
			&BoolWithInverseFlag{
				Name:  "bif",
				Value: true, // this value should fail validation
				Validator: func(i bool) error {
					if i {
						return fmt.Errorf("invalid value")
					}
					return nil
				},
				ValidateDefaults: true,
			},
		},
	}

	r := require.New(t)

	// Default value of flag is 2 which should fail validation
	err := cmd.Run(buildTestContext(t), []string{"foo", "--bif"})
	r.Error(err)
}

func TestFlagValidation(t *testing.T) {
	r := require.New(t)

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
				&Int64Flag{
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
			r.NoError(err)
		} else {
			r.Error(err)
		}
	}
}

func TestBoolInverseFlagValidation(t *testing.T) {
	r := require.New(t)

	cmd := &Command{
		Name: "foo",
		Flags: []Flag{
			&BoolWithInverseFlag{
				Name: "it",
				Validator: func(b bool) error {
					if b {
						return nil
					}
					return fmt.Errorf("not valid")
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo", "--it=false"})
	r.Error(err)
}
