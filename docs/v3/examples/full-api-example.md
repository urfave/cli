---
tags:
  - v3
search:
  boost: 2
---

**Notice**: This is a contrived (functioning) example meant strictly for API
demonstration purposes. Use of one's imagination is encouraged.

<!-- {
  "output": "made it!\nPhew!"
} -->
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
	"slices"

	"github.com/urfave/cli/v3"
)

func init() {
	cli.RootCommandHelpTemplate += "\nCUSTOMIZED: you bet ur muffins\n"
	cli.CommandHelpTemplate += "\nYMMV\n"
	cli.SubcommandHelpTemplate += "\nor something\n"

	cli.HelpFlag = &cli.BoolFlag{Name: "halp"}
	cli.VersionFlag = &cli.BoolFlag{Name: "print-version", Aliases: []string{"V"}}

	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		fmt.Fprintf(w, "best of luck to you\n")
	}
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintf(cmd.Root().Writer, "version=%s\n", cmd.Root().Version)
	}
	cli.OsExiter = func(cmd int) {
		fmt.Fprintf(cli.ErrWriter, "refusing to exit %d\n", cmd)
	}
	cli.ErrWriter = ioutil.Discard
	cli.FlagStringer = func(fl cli.Flag) string {
		return fmt.Sprintf("\t\t%s", fl.Names()[0])
	}
}

type hexWriter struct{}

func (w *hexWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		fmt.Printf("%x", b)
	}
	fmt.Printf("\n")

	return len(p), nil
}

type genericType struct {
	s string
}

func (g *genericType) Set(value string) error {
	g.s = value
	return nil
}

func (g *genericType) String() string {
	return g.s
}

func main() {
	cmd := &cli.Command{
		Name:    "kənˈtrīv",
		Version: "v19.99.0",
		/*Authors: []any{
			&cli.Author{
				Name:  "Example Human",
				Email: "human@example.com",
			},
		},*/
		Copyright: "(c) 1999 Serious Enterprise",
		Usage:     "demonstrate available API",
		UsageText: "contrive - demonstrating the available API",
		ArgsUsage: "[args and such]",
		Commands: []*cli.Command{
			&cli.Command{
				Name:        "doo",
				Aliases:     []string{"do"},
				Category:    "motion",
				Usage:       "do the doo",
				UsageText:   "doo - does the dooing",
				Description: "no really, there is a lot of dooing to be done",
				ArgsUsage:   "[arrgh]",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "forever", Aliases: []string{"forevvarr"}},
				},
				Commands: []*cli.Command{
					&cli.Command{
						Name:   "wop",
						Action: wopAction,
					},
				},
				SkipFlagParsing: false,
				HideHelp:        false,
				Hidden:          false,
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					fmt.Fprintf(cmd.Root().Writer, "--better\n")
				},
				Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
					fmt.Fprintf(cmd.Root().Writer, "brace for impact\n")
					return nil, nil
				},
				After: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Fprintf(cmd.Root().Writer, "did we lose anyone?\n")
					return nil
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cmd.FullName()
					cmd.HasName("wop")
					cmd.Names()
					cmd.VisibleFlags()
					fmt.Fprintf(cmd.Root().Writer, "dodododododoodododddooooododododooo\n")
					if cmd.Bool("forever") {
						cmd.Run(ctx, nil)
					}
					return nil
				},
				OnUsageError: func(ctx context.Context, cmd *cli.Command, err error, isSubcommand bool) error {
					fmt.Fprintf(cmd.Root().Writer, "for shame\n")
					return err
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fancy"},
			&cli.BoolFlag{Value: true, Name: "fancier"},
			&cli.DurationFlag{Name: "howlong", Aliases: []string{"H"}, Value: time.Second * 3},
			&cli.FloatFlag{Name: "howmuch"},
			&cli.IntFlag{Name: "longdistance", Validator: func (t int) error {
				if t < 10 {
					return fmt.Errorf("10 miles isnt long distance!!!!")
				}
				return nil
			}},
			&cli.IntSliceFlag{Name: "intervals"},
			&cli.StringFlag{Name: "dance-move", Aliases: []string{"d"}, Validator: func(move string) error {
				moves := []string{"salsa", "tap", "two-step", "lock-step"}
				if !slices.Contains(moves, move) {
					return fmt.Errorf("Havent learnt %s move yet", move)
				}
				return nil
			}},
			&cli.StringSliceFlag{Name: "names", Aliases: []string{"N"}},
			&cli.UintFlag{Name: "age"},
		},
		EnableShellCompletion: true,
		HideHelp:              false,
		HideVersion:           false,
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			fmt.Fprintf(cmd.Root().Writer, "lipstick\nkiss\nme\nlipstick\nringo\n")
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			fmt.Fprintf(cmd.Root().Writer, "HEEEERE GOES\n")
			return nil, nil
		},
		After: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Fprintf(cmd.Root().Writer, "Phew!\n")
			return nil
		},
		CommandNotFound: func(ctx context.Context, cmd *cli.Command, command string) {
			fmt.Fprintf(cmd.Root().Writer, "Thar be no %q here.\n", command)
		},
		OnUsageError: func(ctx context.Context, cmd *cli.Command, err error, isSubcommand bool) error {
			if isSubcommand {
				return err
			}

			fmt.Fprintf(cmd.Root().Writer, "WRONG: %#v\n", err)
			return nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cli.DefaultAppComplete(ctx, cmd)
			cli.HandleExitCoder(errors.New("not an exit coder, though"))
			cli.ShowAppHelp(cmd)
			cli.ShowCommandHelp(ctx, cmd, "also-nope")
			cli.ShowSubcommandHelp(cmd)
			cli.ShowVersion(cmd)

			fmt.Printf("%#v\n", cmd.Root().Command("doo"))
			if cmd.Bool("infinite") {
				cmd.Root().Run(ctx, []string{"app", "doo", "wop"})
			}

			if cmd.Bool("forevar") {
				cmd.Root().Run(ctx, nil)
			}
			fmt.Printf("%#v\n", cmd.Root().VisibleCategories())
			fmt.Printf("%#v\n", cmd.Root().VisibleCommands())
			fmt.Printf("%#v\n", cmd.Root().VisibleFlags())

			fmt.Printf("%#v\n", cmd.Args().First())
			if cmd.Args().Len() > 0 {
				fmt.Printf("%#v\n", cmd.Args().Get(1))
			}
			fmt.Printf("%#v\n", cmd.Args().Present())
			fmt.Printf("%#v\n", cmd.Args().Tail())

			ec := cli.Exit("ohwell", 86)
			fmt.Fprintf(cmd.Root().Writer, "%d", ec.ExitCode())
			fmt.Printf("made it!\n")
			return ec
		},
		Metadata: map[string]interface{}{
			"layers":          "many",
			"explicable":      false,
			"whatever-values": 19.99,
		},
	}

	if os.Getenv("HEXY") != "" {
		cmd.Writer = &hexWriter{}
		cmd.ErrWriter = &hexWriter{}
	}

	cmd.Run(context.Background(), os.Args)
}

func wopAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Fprintf(cmd.Root().Writer, ":wave: over here, eh\n")
	return nil
}
```
