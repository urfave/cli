---
tags:
  - v2
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
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func init() {
	cli.AppHelpTemplate += "\nCUSTOMIZED: you bet ur muffins\n"
	cli.CommandHelpTemplate += "\nYMMV\n"
	cli.SubcommandHelpTemplate += "\nor something\n"

	cli.HelpFlag = &cli.BoolFlag{Name: "halp"}
	cli.VersionFlag = &cli.BoolFlag{Name: "print-version", Aliases: []string{"V"}}

	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		fmt.Fprintf(w, "best of luck to you\n")
	}
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Fprintf(cCtx.App.Writer, "version=%s\n", cCtx.App.Version)
	}
	cli.OsExiter = func(cCtx int) {
		fmt.Fprintf(cli.ErrWriter, "refusing to exit %d\n", cCtx)
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
	app := &cli.App{
		Name:     "kənˈtrīv",
		Version:  "v19.99.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Example Human",
				Email: "human@example.com",
			},
		},
		Copyright: "(c) 1999 Serious Enterprise",
		HelpName:  "contrive",
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
				Subcommands: []*cli.Command{
					&cli.Command{
						Name:   "wop",
						Action: wopAction,
					},
				},
				SkipFlagParsing:    false,
				HideHelp:           false,
				HideHelpCommand:    false,
				Hidden:             false,
				HelpName:           "doo!",
				BashComplete: func(cCtx *cli.Context) {
					fmt.Fprintf(cCtx.App.Writer, "--better\n")
				},
				Before: func(cCtx *cli.Context) error {
					fmt.Fprintf(cCtx.App.Writer, "brace for impact\n")
					return nil
				},
				After: func(cCtx *cli.Context) error {
					fmt.Fprintf(cCtx.App.Writer, "did we lose anyone?\n")
					return nil
				},
				Action: func(cCtx *cli.Context) error {
					cCtx.Command.FullName()
					cCtx.Command.HasName("wop")
					cCtx.Command.Names()
					cCtx.Command.VisibleFlags()
					fmt.Fprintf(cCtx.App.Writer, "dodododododoodododddooooododododooo\n")
					if cCtx.Bool("forever") {
						cCtx.Command.Run(cCtx)
					}
					return nil
				},
				OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
					fmt.Fprintf(cCtx.App.Writer, "for shame\n")
					return err
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fancy"},
			&cli.BoolFlag{Value: true, Name: "fancier"},
			&cli.DurationFlag{Name: "howlong", Aliases: []string{"H"}, Value: time.Second * 3},
			&cli.Float64Flag{Name: "howmuch"},
			&cli.GenericFlag{Name: "wat", Value: &genericType{}},
			&cli.Int64Flag{Name: "longdistance"},
			&cli.Int64SliceFlag{Name: "intervals"},
			&cli.IntFlag{Name: "distance"},
			&cli.IntSliceFlag{Name: "times"},
			&cli.StringFlag{Name: "dance-move", Aliases: []string{"d"}},
			&cli.StringSliceFlag{Name: "names", Aliases: []string{"N"}},
			&cli.UintFlag{Name: "age"},
			&cli.Uint64Flag{Name: "bigage"},
		},
		EnableBashCompletion: true,
		HideHelp:             false,
		HideHelpCommand:      false,
		HideVersion:          false,
		BashComplete: func(cCtx *cli.Context) {
			fmt.Fprintf(cCtx.App.Writer, "lipstick\nkiss\nme\nlipstick\nringo\n")
		},
		Before: func(cCtx *cli.Context) error {
			fmt.Fprintf(cCtx.App.Writer, "HEEEERE GOES\n")
			return nil
		},
		After: func(cCtx *cli.Context) error {
			fmt.Fprintf(cCtx.App.Writer, "Phew!\n")
			return nil
		},
		CommandNotFound: func(cCtx *cli.Context, command string) {
			fmt.Fprintf(cCtx.App.Writer, "Thar be no %q here.\n", command)
		},
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			if isSubcommand {
				return err
			}

			fmt.Fprintf(cCtx.App.Writer, "WRONG: %#v\n", err)
			return nil
		},
		Action: func(cCtx *cli.Context) error {
			cli.DefaultAppComplete(cCtx)
			cli.HandleExitCoder(errors.New("not an exit coder, though"))
			cli.ShowAppHelp(cCtx)
			cli.ShowCommandCompletions(cCtx, "nope")
			cli.ShowCommandHelp(cCtx, "also-nope")
			cli.ShowCompletions(cCtx)
			cli.ShowSubcommandHelp(cCtx)
			cli.ShowVersion(cCtx)

			fmt.Printf("%#v\n", cCtx.App.Command("doo"))
			if cCtx.Bool("infinite") {
				cCtx.App.Run([]string{"app", "doo", "wop"})
			}

			if cCtx.Bool("forevar") {
				cCtx.App.RunAsSubcommand(cCtx)
			}
			cCtx.App.Setup()
			fmt.Printf("%#v\n", cCtx.App.VisibleCategories())
			fmt.Printf("%#v\n", cCtx.App.VisibleCommands())
			fmt.Printf("%#v\n", cCtx.App.VisibleFlags())

			fmt.Printf("%#v\n", cCtx.Args().First())
			if cCtx.Args().Len() > 0 {
				fmt.Printf("%#v\n", cCtx.Args().Get(1))
			}
			fmt.Printf("%#v\n", cCtx.Args().Present())
			fmt.Printf("%#v\n", cCtx.Args().Tail())

			set := flag.NewFlagSet("contrive", 0)
			nc := cli.NewContext(cCtx.App, set, cCtx)

			fmt.Printf("%#v\n", nc.Args())
			fmt.Printf("%#v\n", nc.Bool("nope"))
			fmt.Printf("%#v\n", !nc.Bool("nerp"))
			fmt.Printf("%#v\n", nc.Duration("howlong"))
			fmt.Printf("%#v\n", nc.Float64("hay"))
			fmt.Printf("%#v\n", nc.Generic("bloop"))
			fmt.Printf("%#v\n", nc.Int64("bonk"))
			fmt.Printf("%#v\n", nc.Int64Slice("burnks"))
			fmt.Printf("%#v\n", nc.Int("bips"))
			fmt.Printf("%#v\n", nc.IntSlice("blups"))
			fmt.Printf("%#v\n", nc.String("snurt"))
			fmt.Printf("%#v\n", nc.StringSlice("snurkles"))
			fmt.Printf("%#v\n", nc.Uint("flub"))
			fmt.Printf("%#v\n", nc.Uint64("florb"))

			fmt.Printf("%#v\n", nc.FlagNames())
			fmt.Printf("%#v\n", nc.IsSet("wat"))
			fmt.Printf("%#v\n", nc.Set("wat", "nope"))
			fmt.Printf("%#v\n", nc.NArg())
			fmt.Printf("%#v\n", nc.NumFlags())
			fmt.Printf("%#v\n", nc.Lineage()[1])
			nc.Set("wat", "also-nope")

			ec := cli.Exit("ohwell", 86)
			fmt.Fprintf(cCtx.App.Writer, "%d", ec.ExitCode())
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
		app.Writer = &hexWriter{}
		app.ErrWriter = &hexWriter{}
	}

	app.Run(os.Args)
}

func wopAction(cCtx *cli.Context) error {
	fmt.Fprintf(cCtx.App.Writer, ":wave: over here, eh\n")
	return nil
}
```
