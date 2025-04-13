package cli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestArgumentsRootCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedIvals  []int64
		expectedUivals []uint64
		expectedFvals  []float64
		errStr         string
	}{
		{
			name:           "set ival",
			args:           []string{"foo", "10"},
			expectedIvals:  []int64{10},
			expectedUivals: []uint64{},
			expectedFvals:  []float64{},
		},
		{
			name:           "set invalid ival",
			args:           []string{"foo", "10.0"},
			expectedIvals:  []int64{},
			expectedUivals: []uint64{},
			expectedFvals:  []float64{},
			errStr:         "strconv.ParseInt: parsing \"10.0\": invalid syntax",
		},
		{
			name:           "set ival uival",
			args:           []string{"foo", "-10", "11"},
			expectedIvals:  []int64{-10},
			expectedUivals: []uint64{11},
			expectedFvals:  []float64{},
		},
		{
			name:           "set ival uival fval",
			args:           []string{"foo", "-12", "14", "10.1"},
			expectedIvals:  []int64{-12},
			expectedUivals: []uint64{14},
			expectedFvals:  []float64{10.1},
		},
		{
			name:           "set ival uival multu fvals",
			args:           []string{"foo", "-13", "12", "10.1", "11.09"},
			expectedIvals:  []int64{-13},
			expectedUivals: []uint64{12},
			expectedFvals:  []float64{10.1, 11.09},
		},
		{
			name:           "set fvals beyond max",
			args:           []string{"foo", "13", "10", "10.1", "11.09", "12.1"},
			expectedIvals:  []int64{13},
			expectedUivals: []uint64{10},
			expectedFvals:  []float64{10.1, 11.09},
			errStr:         "No help topic for '12.1'",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := buildMinimalTestCommand()
			var ivals []int64
			var uivals []uint64
			var fvals []float64
			cmd.Arguments = []Argument{
				&IntArgs{
					Name:        "ia",
					Min:         1,
					Max:         1,
					Destination: &ivals,
				},
				&UintArgs{
					Name:        "uia",
					Min:         1,
					Max:         1,
					Destination: &uivals,
				},
				&FloatArgs{
					Name:        "fa",
					Min:         0,
					Max:         2,
					Destination: &fvals,
				},
			}

			err := cmd.Run(buildTestContext(t), test.args)

			r := require.New(t)

			if test.errStr != "" {
				r.ErrorContains(err, test.errStr)
			} else {
				r.Equal(test.expectedIvals, ivals)
			}
			r.Equal(test.expectedIvals, cmd.IntArgs("ia"))
			r.Equal(test.expectedFvals, cmd.FloatArgs("fa"))
			r.Equal(test.expectedUivals, cmd.UintArgs("uia"))
			/*if test.expectedFvals != nil {
				r.Equal(test.expectedFvals, fvals)
			}*/
		})
	}

	/*
	   cmd.Arguments = append(cmd.Arguments,

	   	&StringArgs{
	   		Name: "sa",
	   	},
	   	&UintArgs{
	   		Name: "ua",
	   		Min:  2,
	   		Max:  1, // max is less than min
	   	},

	   )

	   require.NoError(t, cmd.Run(context.Background(), []string{"foo", "10"}))
	*/
}

func TestArgumentsInvalidType(t *testing.T) {
	cmd := buildMinimalTestCommand()
	cmd.Arguments = []Argument{
		&IntArgs{
			Name: "ia",
			Min:  1,
			Max:  1,
		},
	}
	r := require.New(t)
	r.Nil(cmd.StringArgs("ia"))
	r.Nil(cmd.FloatArgs("ia"))
	r.Nil(cmd.UintArgs("ia"))
	r.Nil(cmd.TimestampArgs("ia"))
	r.Nil(cmd.IntArgs("uia"))
}

func TestArgumentsSubcommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedIval  int64
		expectedSvals []string
		expectedTVals []time.Time
		errStr        string
	}{
		{
			name:   "insuff args",
			args:   []string{"foo", "subcmd", "2006-01-02T15:04:05Z"},
			errStr: "sufficient count of arg sa not provided, given 0 expected 1",
		},
		{
			name:          "set sval and tval",
			args:          []string{"foo", "subcmd", "2006-01-02T15:04:05Z", "fubar"},
			expectedIval:  10,
			expectedTVals: []time.Time{time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)},
			expectedSvals: []string{"fubar"},
		},
		{
			name:          "set sval, tval and ival",
			args:          []string{"foo", "subcmd", "--foo", "100", "2006-01-02T15:04:05Z", "fubar", "some"},
			expectedIval:  100,
			expectedTVals: []time.Time{time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)},
			expectedSvals: []string{"fubar", "some"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := buildMinimalTestCommand()
			var ival int64
			var svals []string
			var tvals []time.Time
			cmd.Commands = []*Command{
				{
					Name: "subcmd",
					Flags: []Flag{
						&IntFlag{
							Name:        "foo",
							Value:       10,
							Destination: &ival,
						},
					},
					Arguments: []Argument{
						&TimestampArgs{
							Name:        "ta",
							Min:         1,
							Max:         1,
							Destination: &tvals,
							Config: TimestampConfig{
								Layouts: []string{time.RFC3339},
							},
						},
						&StringArgs{
							Name:        "sa",
							Min:         1,
							Max:         3,
							Destination: &svals,
						},
					},
				},
			}

			numUsageErrors := 0
			cmd.Commands[0].OnUsageError = func(ctx context.Context, cmd *Command, err error, isSubcommand bool) error {
				numUsageErrors++
				return err
			}

			err := cmd.Run(buildTestContext(t), test.args)

			r := require.New(t)

			if test.errStr != "" {
				r.ErrorContains(err, test.errStr)
				r.Equal(1, numUsageErrors)
			} else {
				if test.expectedSvals != nil {
					r.Equal(test.expectedSvals, svals)
					r.Equal(test.expectedSvals, cmd.Commands[0].StringArgs("sa"))
				}
				if test.expectedTVals != nil {
					r.Equal(test.expectedTVals, tvals)
					r.Equal(test.expectedTVals, cmd.Commands[0].TimestampArgs("ta"))
				}
				r.Equal(test.expectedIval, ival)
			}
		})
	}
}

func TestArgsUsage(t *testing.T) {
	arg := &IntArgs{
		Name: "ia",
		Min:  0,
		Max:  1,
	}
	tests := []struct {
		name     string
		min      int
		max      int
		usage    string
		expected string
	}{
		{
			name:     "optional",
			min:      0,
			max:      1,
			expected: "[ia]",
		},
		{
			name:     "optional",
			min:      0,
			max:      1,
			usage:    "[my optional usage]",
			expected: "[my optional usage]",
		},
		{
			name:     "zero or more",
			min:      0,
			max:      2,
			expected: "[ia ...]",
		},
		{
			name:     "one",
			min:      1,
			max:      1,
			expected: "ia [ia ...]",
		},
		{
			name:     "many",
			min:      2,
			max:      1,
			expected: "ia [ia ...]",
		},
		{
			name:     "many2",
			min:      2,
			max:      0,
			expected: "ia [ia ...]",
		},
		{
			name:     "unlimited",
			min:      2,
			max:      -1,
			expected: "ia [ia ...]",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arg.Min, arg.Max, arg.UsageText = test.min, test.max, test.usage
			require.Equal(t, test.expected, arg.Usage())
		})
	}
}

func TestSingleOptionalArg(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		argValue string
		exp      []string
	}{
		{
			name: "no args",
			args: []string{"foo"},
			exp:  []string{},
		},
		/*{
			name: "no arg with def value",
			args: []string{"foo"},
			exp:  []string{"bar"},
		},*/
		{
			name: "one arg",
			args: []string{"foo", "zbar"},
			exp:  []string{"zbar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := buildMinimalTestCommand()
			var s1 []string
			arg := &StringArgs{
				Min:         0,
				Max:         1,
				Value:       test.argValue,
				Destination: &s1,
			}
			cmd.Arguments = []Argument{
				arg,
			}

			err := cmd.Run(buildTestContext(t), test.args) //
			r := require.New(t)
			r.NoError(err)
			r.Equal(test.exp, s1)
		})
	}
}

func TestUnboundedArgs(t *testing.T) {
	arg := &StringArgs{
		Min: 0,
		Max: -1,
	}
	tests := []struct {
		name      string
		args      []string
		defValues []string
		values    []string
		expected  []string
	}{
		{
			name:     "cmd accepts no args",
			args:     []string{"foo"},
			expected: []string{},
		},
		{
			name:     "cmd uses given args",
			args:     []string{"foo", "bar", "baz"},
			expected: []string{"bar", "baz"},
		},
		{
			name:     "cmd uses default values",
			args:     []string{"foo"},
			expected: []string{},
		},
		{
			name:     "given args override default values",
			args:     []string{"foo", "bar", "baz"},
			values:   []string{"zbar", "zbaz"},
			expected: []string{"bar", "baz"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := buildMinimalTestCommand()
			cmd.Arguments = []Argument{arg}
			arg.Destination = &test.values
			require.NoError(t, cmd.Run(context.Background(), test.args))
			require.Equal(t, test.expected, *arg.Destination)
		})
	}
}
