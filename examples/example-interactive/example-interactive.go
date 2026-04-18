package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "interactive-demo",
		Usage: "A demonstration of interactive mode in urfave/cli",
		Description: `This example demonstrates how to use the interactive mode
to prompt users for missing parameters. Use --interactive or -i flag
to enable interactive prompting.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Enable interactive mode for missing parameters",
			},
			&cli.InteractiveStringFlag{
				StringFlag: cli.StringFlag{
					Name:  "name",
					Value: "Guest",
					Usage: "Your name",
				},
				Prompt:   "Enter your name",
				Required: true,
			},
			&cli.InteractiveIntFlag{
				Int64Flag: cli.Int64Flag{
					Name:  "age",
					Value: 18,
					Usage: "Your age",
				},
				Prompt:   "Enter your age",
				Required: true,
			},
			&cli.InteractiveStringFlag{
				StringFlag: cli.StringFlag{
					Name:  "email",
					Usage: "Your email address",
				},
				Prompt:   "Enter your email (optional)",
				Required: false,
			},
			&cli.InteractiveBoolFlag{
				BoolFlag: cli.BoolFlag{
					Name:  "subscribe",
					Value: false,
					Usage: "Subscribe to newsletter",
				},
				Prompt: "Do you want to subscribe to our newsletter",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Println("\n=== User Profile ===")
			fmt.Printf("Name: %s\n", cmd.String("name"))
			fmt.Printf("Age: %d\n", cmd.Int("age"))
			if email := cmd.String("email"); email != "" {
				fmt.Printf("Email: %s\n", email)
			} else {
				fmt.Println("Email: Not provided")
			}
			fmt.Printf("Subscribed: %v\n", cmd.Bool("subscribe"))

			if cmd.Bool("subscribe") {
				fmt.Println("\nThank you for subscribing!")
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a new project interactively",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "interactive",
						Aliases: []string{"i"},
						Usage:   "Enable interactive mode",
					},
					&cli.InteractiveStringFlag{
						StringFlag: cli.StringFlag{
							Name:  "project-name",
							Usage: "Name of the project",
						},
						Prompt:   "Enter project name",
						Required: true,
					},
					&cli.InteractiveStringFlag{
						StringFlag: cli.StringFlag{
							Name:  "description",
							Value: "A new project",
							Usage: "Project description",
						},
						Prompt:   "Enter project description",
						Required: false,
					},
					&cli.InteractiveStringFlag{
						StringFlag: cli.StringFlag{
							Name:  "language",
							Value: "go",
							Usage: "Programming language",
						},
						Prompt:   "Enter programming language",
						Required: false,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("\n=== Project Created ===")
					fmt.Printf("Project Name: %s\n", cmd.String("project-name"))
					fmt.Printf("Description: %s\n", cmd.String("description"))
					fmt.Printf("Language: %s\n", cmd.String("language"))
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
