package cli

import (
	"testing"
)

func TestFlagExGroup(t *testing.T) {

	a := &App{
		FlagExGroups: []FlagExGroup{
			{
				Flags: [][]Flag{
					[]Flag{
						&IntFlag{
							Name: "i",
						},
						&StringFlag{
							Name: "s",
						},
					},
					[]Flag{
						&Int64Flag{
							Name:    "t",
							Aliases: []string{"ai"},
						},
					},
				},
			},
		},
	}

	err := a.Run([]string{"foo"})
	if err != nil {
		t.Error(err)
	}

	err = a.Run([]string{"foo", "--i", "10"})
	if err != nil {
		t.Error(err)
	}

	err = a.Run([]string{"foo", "--i", "11", "--ai", "12"})
	if err == nil {
		t.Error("Expected mutual exclusion error")
	} else if _, ok := err.(*mutuallyExclusiveGroup); !ok {
		t.Errorf("Got invalid error %v", err)
	}

	a.FlagExGroups[0].Required = true

	err = a.Run([]string{"foo"})
	if err == nil {
		t.Error("Required flags error")
	} else if _, ok := err.(*mutuallyExclusiveGroupRequiredFlag); !ok {
		t.Errorf("Got invalid error %v", err)
	}

	err = a.Run([]string{"foo", "--i", "10"})
	if err != nil {
		t.Error(err)
	}

	err = a.Run([]string{"foo", "--i", "11", "--ai", "12"})
	if err == nil {
		t.Error("Expected mutual exclusion error")
	} else if _, ok := err.(*mutuallyExclusiveGroup); !ok {
		t.Errorf("Got invalid error %v", err)
	}
}
