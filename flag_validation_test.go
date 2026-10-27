package cli

import (
	"context"
	"fmt"
	"io"
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

func TestMultiValueFlagValidation(t *testing.T) {
	minTwo := func(v []string) error {
		if len(v) < 2 {
			return fmt.Errorf("need at least 2 values, got %q", v)
		}
		return nil
	}
	maxTwo := func(v []string) error {
		if len(v) > 2 {
			return fmt.Errorf("need at most 2 values, got %q", v)
		}
		return nil
	}
	needAB := func(m map[string]string) error {
		if m["a"] == "" || m["b"] == "" {
			return fmt.Errorf("need keys a and b, got %v", m)
		}
		return nil
	}

	testCases := []struct {
		name    string
		flag    Flag
		env     string
		args    []string
		wantErr string
		want    any
	}{
		{
			name: "slice given once",
			flag: &StringSliceFlag{Name: "l", Validator: minTwo},
			args: []string{"-l", "a,b"},
			want: []string{"a", "b"},
		},
		{
			name: "slice given twice",
			flag: &StringSliceFlag{Name: "l", Validator: minTwo},
			args: []string{"-l", "a", "-l", "b"},
			want: []string{"a", "b"},
		},
		{
			name:    "slice with too few values",
			flag:    &StringSliceFlag{Name: "l", Validator: minTwo},
			args:    []string{"-l", "a"},
			wantErr: `invalid value "a" for flag -l: need at least 2 values, got ["a"]`,
		},
		{
			name:    "slice with too many values",
			flag:    &StringSliceFlag{Name: "list", Aliases: []string{"l"}, Validator: maxTwo},
			args:    []string{"-l", "a", "--list", "b,c"},
			wantErr: `invalid value ["a" "b,c"] for flag -list: need at most 2 values, got ["a" "b" "c"]`,
		},
		{
			name: "slice from env",
			flag: &StringSliceFlag{Name: "l", Validator: minTwo, Sources: EnvVars("TEST_LIST")},
			env:  "a,b",
			want: []string{"a", "b"},
		},
		{
			name:    "slice from env with too few values",
			flag:    &StringSliceFlag{Name: "l", Validator: minTwo, Sources: EnvVars("TEST_LIST")},
			env:     "a",
			wantErr: `need at least 2 values, got ["a"]`,
		},
		{
			name: "map given twice",
			flag: &StringMapFlag{Name: "m", Validator: needAB},
			args: []string{"-m", "a=1", "-m", "b=2"},
			want: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:    "map missing a key",
			flag:    &StringMapFlag{Name: "m", Validator: needAB},
			args:    []string{"-m", "a=1"},
			wantErr: `invalid value "a=1" for flag -m: need keys a and b, got map[a:1]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.env != "" {
				t.Setenv("TEST_LIST", tc.env)
			}
			var got any
			cmd := &Command{
				Name:      "foo",
				Writer:    io.Discard,
				ErrWriter: io.Discard,
				Flags:     []Flag{tc.flag},
				Action: func(_ context.Context, cmd *Command) error {
					got = cmd.Value(tc.flag.Names()[0])
					return nil
				},
			}

			err := cmd.Run(buildTestContext(t), append([]string{"foo"}, tc.args...))
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestMultiValueFlagValidatorRunsOnce(t *testing.T) {
	var calls [][]string
	cmd := &Command{
		Name: "foo",
		Flags: []Flag{&StringSliceFlag{Name: "l", Validator: func(v []string) error {
			calls = append(calls, v)
			return nil
		}}},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"foo", "-l", "a", "-l", "b", "-l", "c"}))
	require.Equal(t, [][]string{{"a", "b", "c"}}, calls)
}
