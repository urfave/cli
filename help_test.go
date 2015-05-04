package cli_test

import (
	"bytes"
	"testing"

	"github.com/codegangsta/cli"
)

func Test_ShowAppHelp_NoAuthor(t *testing.T) {
	output := new(bytes.Buffer)
	app := cli.NewApp()
	app.Writer = output

	c := cli.NewContext(app, nil, nil)

	cli.ShowAppHelp(c)

	if bytes.Index(output.Bytes(), []byte("AUTHOR(S):")) != -1 {
		t.Errorf("expected\n%snot to include %s", output.String(), "AUTHOR(S):")
	}
}
