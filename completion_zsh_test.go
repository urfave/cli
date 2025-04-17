package cli_test

import (
	"context"
	"fmt"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"
)

var devices = []string{"Pixel 7 API 34", "iPhone 12 mini", "iPhone 15"}

func exampleAction(ctx context.Context, c *cli.Command) error {
	fmt.Printf("command %#v called with args: %#v\n", c.Name, c.Args().Slice())
	return nil
}

func makeExampleApp() *cli.Command {
	return &cli.Command{
		Name:                  "emu-cli",
		Usage:                 "Manage android emulators with ease",
		EnableShellCompletion: true,
		HideHelpCommand:       true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "do not print invocations of subprocesses",
				Action: func(ctx context.Context, c *cli.Command, value bool) error {
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Start a single device",
				Action: exampleAction,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "fast",
						Usage: "Run device quickly",
					},
					&cli.BoolFlag{
						Name:  "slow",
						Usage: "Don't hurry up too much",
					},
				},
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					for _, device := range devices {
						fmt.Println(device)
					}
				},
			},
			// TODO: Not expressible in urfave/cli
			{
				Name:   "runall",
				Usage:  "Start many devices",
				Action: exampleAction,
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					for _, device := range devices {
						fmt.Println(device)
					}
				},
			},
			// TODO: Not expressible in urfave/cli
			{
				Name:   "kill",
				Usage:  "Kill a single device",
				Action: exampleAction,
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					for _, device := range devices {
						fmt.Println(device)
					}
				},
			},
			{
				Name:   "create",
				Usage:  "Create a new device",
				Action: exampleAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "os",
						Usage: "OS of the device",
						// TODO: func ShellComplete
					},
					&cli.StringFlag{
						Name:  "os-version",
						Usage: "OS image version",
						// TODO: func ShellComplete
					},
					&cli.StringFlag{
						Name:  "frame",
						Usage: "Frame of the device",
						// TODO: func ShellComplete
					},
				},
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					androidCompletions := []string{"Android 14 (API 33)", "Android 14 (API 33) Play Store", "Android 15 (API 34)"}
					iosCompletions := []string{"iOS 15", "iOS 16", "iOS 17"}

					completions := make([]string, 0)
					completions = append(completions, androidCompletions...)
					completions = append(completions, iosCompletions...)
				},
			},
		},
		CommandNotFound: func(ctx context.Context, c *cli.Command, command string) {
			log.Printf("invalid command '%s'", command)
		},
	}
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

func ExampleCommand_Run_completeCommands_1() {
	cmd := makeExampleApp()
	os.Args = []string{"emu-cli", "", "--generate-shell-completion"}
	os.Setenv("SHELL", "/usr/bin/zsh")

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// run:Start a single device
	// runall:Start many devices
	// kill:Kill a single device
	// create:Create a new device
}

func ExampleCommand_Run_completeCommands_2() {
	cmd := makeExampleApp()
	os.Args = []string{"emu-cli", "r", "--generate-shell-completion"}
	os.Setenv("SHELL", "/usr/bin/zsh")

	// Note: this output is the same as the test above. The actual matching
	// is done by the shell, not by us!

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// run:Start a single device
	// runall:Start many devices
	// kill:Kill a single device
	// create:Create a new device
}

func ExampleCommand_Run_completeFlags() {
	cmd := makeExampleApp()
	os.Args = []string{"emu-cli", "run", "--", "--generate-shell-completion"}
	os.Setenv("SHELL", "/usr/bin/zsh")

	// Note: this output is the same as the test above. The actual matching
	// is done by the shell, not by us!

	_ = cmd.Run(context.Background(), os.Args)
	// Output:
	// --fast:Run device quickly
	// --slow:Don't hurry up too much
}
