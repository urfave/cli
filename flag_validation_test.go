package cli

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagDefaultValidation(t *testing.T) {

	cmd := &Command{
		Name: "foo",
		Flags: []Flag{
			&IntFlag{
				Name:  "if",
				Value: 2,
				Validator: func(i int64) error {
					if (i >= 3 && i <= 10) || (i >= 20 && i <= 24) {
						return nil
					}
					return fmt.Errorf("Value %d not in range [3,10] or [20,24]", i)
				},
			},
		},
	}

	r := require.New(t)

	// Default value of flag is 2 which should fail validation
	err := cmd.Run(context.Background(), []string{"foo", "--if", "5"})
	r.Error(err)
}

func TestFlagValidation(t *testing.T) {

	r := require.New(t)

	testCases := []struct {
		name        string
		arg         string
		errExpected bool
	}{
		/*{
			name:        "first range less than min",
			arg:         "2",
			errExpected: true,
		},*/
		{
			name: "first range min",
			arg:  "3",
		},
		/*{
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
		},*/
	}

	for _, testCase := range testCases {
		cmd := &Command{
			Name: "foo",
			Flags: []Flag{
				&IntFlag{
					Name:  "if",
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

		err := cmd.Run(context.Background(), []string{"foo", "--if", testCase.arg})
		if !testCase.errExpected {
			r.NoError(err)
		} else {
			r.Error(err)
		}
	}
}
