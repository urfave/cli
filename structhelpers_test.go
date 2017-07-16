package cli_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"gitlab.com/ayufan/golang-cli-helpers"
	"testing"
)

func TestStructHelperForArray(t *testing.T) {
	app := cli.NewApp()

	m := &cmd{}
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "hi",
			Action: m.Run,
			Flags:  clihelpers.GetFlagsFromStruct(m),
		},
	}

	if err := app.Run([]string{"main.go", "hi", "--test", "1", "--test", "2"}); err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, []string{"1", "2"}, m.Test, "Expect results to match passed flags")
}

type cmd struct {
	Test []string `short:"t" long:"test" description:"Hi"`
}

func (m *cmd) Run(c *cli.Context) error {
	return nil
}
