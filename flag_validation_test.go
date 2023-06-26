package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRangeValidation(t *testing.T) {

	cmd := &Command{
		Name: "foo",
		Flags: []Flag{
			&IntFlag{
				Name: "if",
				Validator: ValidationChainAny[int64](
					RangeInclusive[int64](3, 10),
					RangeInclusive[int64](20, 24),
				),
			},
		},
	}

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
		err := cmd.Run(context.Background(), []string{"foo", "-if", testCase.arg})
		if !testCase.errExpected {
			r.NoError(err)
		} else {
			r.Error(err)
		}
	}
}
