package altsrc_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func ExampleApp_Run_yamlFileLoaderDuration() {
	execServe := func(c *cli.Context) error {
		keepaliveInterval := c.Duration("keepalive-interval")
		fmt.Printf("keepalive %s\n", keepaliveInterval)
		return nil
	}

	fileExists := func(filename string) bool {
		stat, _ := os.Stat(filename)
		return stat != nil
	}

	// initConfigFileInputSource is like altsrc.InitInputSourceWithContext and altsrc.NewYamlSourceFromFlagFunc, but checks
	// if the config flag is exists and only loads it if it does. If the flag is set and the file exists, it fails.
	initConfigFileInputSource := func(configFlag string, flags []cli.Flag) cli.BeforeFunc {
		return func(context *cli.Context) error {
			configFile := context.String(configFlag)
			if context.IsSet(configFlag) && !fileExists(configFile) {
				return fmt.Errorf("config file %s does not exist", configFile)
			} else if !context.IsSet(configFlag) && !fileExists(configFile) {
				return nil
			}
			inputSource, err := altsrc.NewYamlSourceFromFile(configFile)
			if err != nil {
				return err
			}
			return altsrc.ApplyInputSourceValues(context, inputSource, flags)
		}
	}

	flagsServe := []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			EnvVars:     []string{"CONFIG_FILE"},
			Value:       "../testdata/empty.yml",
			DefaultText: "../testdata/empty.yml",
			Usage:       "config file",
		},
		altsrc.NewDurationFlag(
			&cli.DurationFlag{
				Name:    "keepalive-interval",
				Aliases: []string{"k"},
				EnvVars: []string{"KEEPALIVE_INTERVAL"},
				Value:   45 * time.Second,
				Usage:   "interval of keepalive messages",
			},
		),
	}

	cmdServe := &cli.Command{
		Name:      "serve",
		Usage:     "Run the server",
		UsageText: "serve [OPTIONS..]",
		Action:    execServe,
		Flags:     flagsServe,
		Before:    initConfigFileInputSource("config", flagsServe),
	}

	c := &cli.App{
		Name:                   "cmd",
		HideVersion:            true,
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			cmdServe,
		},
	}

	if err := c.Run([]string{"cmd", "serve", "--config", "../testdata/empty.yml"}); err != nil {
		log.Fatal(err)
	}

	// Output:
	// keepalive 45s
}
