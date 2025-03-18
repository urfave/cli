package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newCommand() *Command {
	return &Command{
		MutuallyExclusiveFlags: []MutuallyExclusiveFlags{
			{
				Flags: [][]Flag{
					{
						&IntFlag{
							Name: "i",
						},
						&StringFlag{
							Name: "s",
						},
						&BoolWithInverseFlag{
							Name: "b",
						},
					},
					{
						&IntFlag{
							Name:    "t",
							Aliases: []string{"ai"},
						},
					},
				},
			},
		},
	}
}

func TestFlagMutuallyExclusiveFlags(t *testing.T) {
	cmd := newCommand()

	err := cmd.Run(buildTestContext(t), []string{"foo"})
	assert.NoError(t, err)

	cmd = newCommand()
	err = cmd.Run(buildTestContext(t), []string{"foo", "--i", "10"})
	assert.NoError(t, err)

	cmd = newCommand()
	err = cmd.Run(buildTestContext(t), []string{"foo", "--i", "11", "--ai", "12"})
	if err == nil {
		t.Error("Expected mutual exclusion error")
	} else if err1, ok := err.(*mutuallyExclusiveGroup); !ok {
		t.Errorf("Got invalid error %v", err)
	} else if !strings.Contains(err1.Error(), "option i cannot be set along with option ai") {
		t.Logf("Invalid error string %v", err1)
	}

	cmd = newCommand()
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
	assert.NoError(t, err)

	err = cmd.Run(buildTestContext(t), []string{"foo", "--i", "11", "--ai", "12"})
	if err == nil {
		t.Error("Expected mutual exclusion error")
	} else if err1, ok := err.(*mutuallyExclusiveGroup); !ok {
		t.Errorf("Got invalid error %v", err)
	} else if !strings.Contains(err1.Error(), "option i cannot be set along with option ai") {
		t.Logf("Invalid error string %v", err1)
	}
}
