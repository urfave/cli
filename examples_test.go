package cli_test

import (
	"context"
	"fmt"
	"net/mail"
	"os"
	"time"

	// Alias the package import to make the examples runnable on pkg.go.dev.
	//
	// See issue #1811.
	cli "github.com/urfave/cli/v3"
)

func ExampleCommand_Run() {
	// Declare a command
	cmd := &cli.Command{
		Name: "greet",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: "pat", Usage: "a name to say"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			fmt.Printf("Hello %[1]v\n", cmd.String("name"))
			return nil
		},
		Authors: []any{
			&mail.Address{Name: "Oliver Allen", Address: "oliver@toyshop.example.com"},
			"gruffalo@soup-world.example.org",
		},
		Version: "v0.13.12",
	}

	// Simulate the command line arguments
	os.Args = []string{"greet", "--name", "Jeremy"}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		// do something with unhandled errors
		fmt.Fprintf(os.Stderr, "Unhandled error: %[1]v\n", err)
		os.Exit(86)
	}
	// Output:
	// Hello Jeremy
}

func ExampleCommand_Run_subcommand() {
	cmd := &cli.Command{
		Name: "say",
		Commands: []*cli.Command{
			{
				Name:        "hello",
				Aliases:     []string{"hi"},
				Usage:       "use it to see a description",
				Description: "This is how we describe hello the function",
				Commands: []*cli.Command{
					{
						Name:        "english",
						Aliases:     []string{"en"},
						Usage:       "sends a greeting in english",
						Description: "greets someone in english",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Value: "Bob",
								Usage: "Name of the person to greet",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							fmt.Println("Hello,", cmd.String("name"))
							return nil
						},
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate the command line arguments
	os.Args = []string{"say", "hi", "english", "--name", "Jeremy"}

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// Hello, Jeremy
}

func ExampleCommand_Run_appHelp() {
	cmd := &cli.Command{
		Name:        "greet",
		Version:     "0.1.0",
		Description: "This is how we describe greet the app",
		Authors: []any{
			&mail.Address{Name: "Harrison", Address: "harrison@lolwut.example.com"},
			"Oliver Allen  <oliver@toyshop.example.com>",
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
		},
		Commands: []*cli.Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				ArgsUsage:   "[arguments...]",
				Action: func(context.Context, *cli.Command) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate the command line arguments
	os.Args = []string{"greet", "help"}

	_ = cmd.Run(ctx, os.Args)
	// Output:
	// NAME:
	//    greet - A new cli application
	//
	// USAGE:
	//    greet [global options] [command [command options]] [arguments...]
	//
	// VERSION:
	//    0.1.0
	//
	// DESCRIPTION:
	//    This is how we describe greet the app
	//
	// AUTHORS:
	//    "Harrison" <harrison@lolwut.example.com>
	//    Oliver Allen  <oliver@toyshop.example.com>
	//
	// COMMANDS:
	//    describeit, d  use it to see a description
	//    help, h        Shows a list of commands or help for one command
	//
	// GLOBAL OPTIONS:
	//    --name value   a name to say (default: "bob")
	//    --help, -h     show help
	//    --version, -v  print the version
}

func ExampleCommand_Run_commandHelp() {
	cmd := &cli.Command{
		Name: "greet",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: "pat", Usage: "a name to say"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			fmt.Fprintf(cmd.Root().Writer, "hello to %[1]q\n", cmd.String("name"))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				ArgsUsage:   "[arguments...]",
				Action: func(context.Context, *cli.Command) error {
					fmt.Println("i like to describe things")
					return nil
				},
			},
		},
	}

	// Simulate the command line arguments
	os.Args = []string{"greet", "h", "describeit"}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// NAME:
	//    greet describeit - use it to see a description
	//
	// USAGE:
	//    greet describeit [command [command options]] [arguments...]
	//
	// DESCRIPTION:
	//    This is how we describe describeit the function
	//
	// COMMANDS:
	//    help, h  Shows a list of commands or help for one command
	//
	// OPTIONS:
	//    --help, -h  show help
}

func ExampleCommand_Run_noAction() {
	cmd := &cli.Command{Name: "greet"}

	// Simulate the command line arguments
	os.Args = []string{"greet"}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// NAME:
	//    greet - A new cli application
	//
	// USAGE:
	//    greet [global options] [command [command options]] [arguments...]
	//
	// COMMANDS:
	//    help, h  Shows a list of commands or help for one command
	//
	// GLOBAL OPTIONS:
	//    --help, -h  show help
}

func ExampleCommand_Run_subcommandNoAction() {
	cmd := &cli.Command{
		Name: "greet",
		Commands: []*cli.Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				ArgsUsage:   "[arguments...]",
				Description: "This is how we describe describeit the function",
			},
		},
	}

	// Simulate the command line arguments
	os.Args = []string{"greet", "describeit"}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// NAME:
	//    greet describeit - use it to see a description
	//
	// USAGE:
	//    greet describeit [command [command options]] [arguments...]
	//
	// DESCRIPTION:
	//    This is how we describe describeit the function
	//
	// OPTIONS:
	//    --help, -h  show help
}

func ExampleCommand_Run_shellComplete_bash_withShortFlag() {
	cmd := &cli.Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "other",
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name:    "xyz",
				Aliases: []string{"x"},
			},
		},
	}

	// Simulate a bash environment and command line arguments
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "-", "--generate-shell-completion"}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// --other
	// --xyz
	// --help
}

func ExampleCommand_Run_shellComplete_bash_withLongFlag() {
	cmd := &cli.Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "other",
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name:    "xyz",
				Aliases: []string{"x"},
			},
			&cli.StringFlag{
				Name: "some-flag,s",
			},
			&cli.StringFlag{
				Name: "similar-flag",
			},
		},
	}

	// Simulate a bash environment and command line arguments
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--s", "--generate-shell-completion"}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// --some-flag
	// --similar-flag
}

func ExampleCommand_Run_shellComplete_bash_withMultipleLongFlag() {
	cmd := &cli.Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "int-flag",
				Aliases: []string{"i"},
			},
			&cli.StringFlag{
				Name:    "string",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name: "string-flag-2",
			},
			&cli.StringFlag{
				Name: "similar-flag",
			},
			&cli.StringFlag{
				Name: "some-flag",
			},
		},
	}

	// Simulate a bash environment and command line arguments
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--st", "--generate-shell-completion"}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// --string
	// --string-flag-2
}

func ExampleCommand_Run_shellComplete_bash() {
	cmd := &cli.Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(context.Context, *cli.Command) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			}, {
				Name:        "next",
				Usage:       "next example",
				Description: "more stuff to see when generating shell completion",
				Action: func(context.Context, *cli.Command) error {
					fmt.Printf("the next example")
					return nil
				},
			},
		},
	}

	// Simulate a bash environment and command line arguments
	os.Setenv("SHELL", "bash")
	os.Args = []string{"greet", "--generate-shell-completion"}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// describeit
	// next
	// help
}

func ExampleCommand_Run_shellComplete_zsh() {
	cmd := &cli.Command{
		Name:                  "greet",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "describeit",
				Aliases:     []string{"d"},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(context.Context, *cli.Command) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			}, {
				Name:        "next",
				Usage:       "next example",
				Description: "more stuff to see when generating bash completion",
				Action: func(context.Context, *cli.Command) error {
					fmt.Printf("the next example")
					return nil
				},
			},
		},
	}

	// Simulate a zsh environment and command line arguments
	os.Args = []string{"greet", "--generate-shell-completion"}
	os.Setenv("SHELL", "/usr/bin/zsh")

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// describeit:use it to see a description
	// next:next example
	// help:Shows a list of commands or help for one command
}

func ExampleCommand_Run_sliceValues() {
	cmd := &cli.Command{
		Name: "multi_values",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{Name: "stringSlice"},
			&cli.FloatSliceFlag{Name: "float64Slice"},
			&cli.IntSliceFlag{Name: "intSlice"},
		},
		HideHelp: true,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			for i, v := range cmd.FlagNames() {
				fmt.Printf("%d-%s %#v\n", i, v, cmd.Value(v))
			}
			err := ctx.Err()
			fmt.Println("error:", err)
			return err
		},
	}

	// Simulate command line arguments
	os.Args = []string{
		"multi_values",
		"--stringSlice", "parsed1,parsed2", "--stringSlice", "parsed3,parsed4",
		"--float64Slice", "13.3,14.4", "--float64Slice", "15.5,16.6",
		"--intSlice", "13,14", "--intSlice", "15,16",
	}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// 0-float64Slice []float64{13.3, 14.4, 15.5, 16.6}
	// 1-intSlice []int64{13, 14, 15, 16}
	// 2-stringSlice []string{"parsed1", "parsed2", "parsed3", "parsed4"}
	// error: <nil>
}

func ExampleCommand_Run_mapValues() {
	cmd := &cli.Command{
		Name: "multi_values",
		Flags: []cli.Flag{
			&cli.StringMapFlag{Name: "stringMap"},
		},
		HideHelp: true,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			for i, v := range cmd.FlagNames() {
				fmt.Printf("%d-%s %#v\n", i, v, cmd.StringMap(v))
			}
			fmt.Printf("notfound %#v\n", cmd.StringMap("notfound"))
			err := ctx.Err()
			fmt.Println("error:", err)
			return err
		},
	}

	// Simulate command line arguments
	os.Args = []string{
		"multi_values",
		"--stringMap", "parsed1=parsed two", "--stringMap", "parsed3=",
	}

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// 0-stringMap map[string]string{"parsed1":"parsed two", "parsed3":""}
	// notfound map[string]string(nil)
	// error: <nil>
}

func ExampleBoolWithInverseFlag() {
	flagWithInverse := &cli.BoolWithInverseFlag{
		BoolFlag: &cli.BoolFlag{
			Name: "env",
		},
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			flagWithInverse,
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			if flagWithInverse.IsSet() {
				if flagWithInverse.Value() {
					fmt.Println("env is set")
				} else {
					fmt.Println("no-env is set")
				}
			}

			return nil
		},
	}

	_ = cmd.Run(context.Background(), []string{"prog", "--no-env"})
	_ = cmd.Run(context.Background(), []string{"prog", "--env"})

	fmt.Println("flags:", len(flagWithInverse.Flags()))

	// Output:
	// no-env is set
	// env is set
	// flags: 2
}

func ExampleCommand_Suggest() {
	cmd := &cli.Command{
		Name:                          "greet",
		ErrWriter:                     os.Stdout,
		Suggest:                       true,
		HideHelp:                      false,
		HideHelpCommand:               true,
		CustomRootCommandHelpTemplate: "(this space intentionally left blank)\n",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: "squirrel", Usage: "a name to say"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			fmt.Printf("Hello %v\n", cmd.String("name"))
			return nil
		},
	}

	if cmd.Run(context.Background(), []string{"greet", "--nema", "chipmunk"}) == nil {
		fmt.Println("Expected error")
	}
	// Output:
	// Incorrect Usage: flag provided but not defined: -nema
	//
	// Did you mean "--name"?
	//
	// (this space intentionally left blank)
}

func ExampleCommand_Suggest_command() {
	cmd := &cli.Command{
		ErrWriter: os.Stdout,
		Name:      "greet",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: "squirrel", Usage: "a name to say"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			fmt.Printf("Hello %v\n", cmd.String("name"))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:               "neighbors",
				HideHelp:           true,
				HideHelpCommand:    true,
				Suggest:            true,
				CustomHelpTemplate: "(this space intentionally left blank)\n",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "smiling"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					if cmd.Bool("smiling") {
						fmt.Println("😀")
					}
					fmt.Println("Hello, neighbors")
					return nil
				},
			},
		},
	}

	if cmd.Run(context.Background(), []string{"greet", "neighbors", "--sliming"}) == nil {
		fmt.Println("Expected error")
	}
	// Output:
	// Incorrect Usage: flag provided but not defined: -sliming
	//
	// Did you mean "--smiling"?
}
