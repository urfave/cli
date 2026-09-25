// Command hello-i18n adds a German catalogue to hello. Run it with
// LANG=de_DE.UTF-8 and an extra argument to see a translated error.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v4"
	"github.com/urfave/cli/v4/i18n/locale/de"
)

func main() {
	cli.Main(context.Background(), &cli.Command{
		Name:      "hello",
		Usage:     "print a greeting",
		Localizer: cli.SelectLocalizer(os.LookupEnv, de.Catalog),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Fprintln(cmd.Out(), "Hallo")
			return nil
		},
	})
}
