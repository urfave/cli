package cli

import (
	"strings"
	"testing"
)

func TestFlagMutuallyExclusiveFlags(t *testing.T) {
	cmd := &Command{
		MutuallyExclusiveFlags: []MutuallyExclusiveFlags{
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
						&IntFlag{
							Name:    "t",
							Aliases: []string{"ai"},
						},
					},
				},
			},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"foo"})
	if err != nil {
		t.Error(err)
	}

	err = cmd.Run(buildTestContext(t), []string{"foo", "--i", "10"})
	if err != nil {
		t.Error(err)
	}

	err = cmd.Run(buildTestContext(t), []string{"foo", "--i", "11", "--ai", "12"})
	if err == nil {
		t.Error("Expected mutual exclusion error")
	} else if err1, ok := err.(*mutuallyExclusiveGroup); !ok {
		t.Errorf("Got invalid error %v", err)
	} else if !strings.Contains(err1.Error(), "option i cannot be set along with option ai") {
		t.Errorf("Invalid error string %v", err1)
	}

	cmd.MutuallyExclusiveFlags[0].Required = true

	err = cmd.Run(buildTestContext(t), []string{"foo"})
	if err == nil {
		t.Error("Required flags error")
	} else if err1, ok := err.(*mutuallyExclusiveGroupRequiredFlag); !ok {
		t.Errorf("Got invalid error %v", err)
	} else if !strings.Contains(err1.Error(), "one of") {
		t.Errorf("Invalid error string %v", err1)
	}

	err = cmd.Run(buildTestContext(t), []string{"foo", "--i", "10"})
	if err != nil {
		t.Error(err)
	}

	err = cmd.Run(buildTestContext(t), []string{"foo", "--i", "11", "--ai", "12"})
	if err == nil {
		t.Error("Expected mutual exclusion error")
	} else if err1, ok := err.(*mutuallyExclusiveGroup); !ok {
		t.Errorf("Got invalid error %v", err)
	} else if !strings.Contains(err1.Error(), "option i cannot be set along with option ai") {
		t.Errorf("Invalid error string %v", err1)
	}
}
