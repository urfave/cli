package altsrc_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
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

func TestYamlFileStringSlice(t *testing.T) {
	_ = ioutil.WriteFile("current.yaml", []byte(`top:
  test: ["s1", "s2"]`), 0666)
	defer os.Remove("current.yaml")

	testFlag := []cli.Flag{
		&altsrc.StringFlag{StringFlag: &cli.StringFlag{Name: "conf"}},
		&altsrc.StringSliceFlag{StringSliceFlag: &cli.StringSliceFlag{Name: "top.test", EnvVars: []string{"THE_TEST"}}},
	}
	app := &cli.App{}
	app.Before = altsrc.InitInputSourceWithContext(testFlag, altsrc.NewYamlSourceFromFlagFunc("conf"))
	app.Action = func(c *cli.Context) error {
		if c.IsSet("top.test") {
			return nil
		} else {
			return errors.New("top.test is not set")
		}
	}
	app.Flags = append(app.Flags, testFlag...)

	test := []string{"testApp", "--conf", "current.yaml"}
	if err := app.Run(test); err != nil {
		t.Error(err)
	}
}

func TestYamlFileUint64(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		err   bool
	}{
		{
			"top.test",
			`top: 
  test: 100`,
			false,
		},
		{
			"test",
			"test: ",
			false,
		},
		{
			"test",
			"test: 100", //int
			false,
		},
		{
			"test",
			"test: -100", //int
			true,
		},
		{
			"test",
			"test: 9223372036854775807", //int
			false,
		},
		{
			"test",
			"test: 9223372036854775808", //uintt64
			false,
		},
		{
			"test",
			"test: 19223372036854775808", //float64
			true,
		},
	}

	for i, test := range tests {
		_ = ioutil.WriteFile("current.yaml", []byte(test.entry), 0666)
		defer os.Remove("current.yaml")

		testFlag := []cli.Flag{
			&altsrc.StringFlag{StringFlag: &cli.StringFlag{Name: "conf"}},
			&altsrc.Uint64Flag{Uint64Flag: &cli.Uint64Flag{Name: test.name}},
		}
		app := &cli.App{}
		app.Flags = append(app.Flags, testFlag...)
		app.Before = altsrc.InitInputSourceWithContext(testFlag, altsrc.NewYamlSourceFromFlagFunc("conf"))

		appCmd := []string{"testApp", "--conf", "current.yaml"}
		err := app.Run(appCmd)
		if result := err != nil; result != test.err {
			t.Error(i, "testcast: expect error but", err)
		}
	}
}

func TestYamlFileUint(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		err   bool
	}{
		{
			"top.test",
			`top: 
  test: 100`,
			false,
		},
		{
			"test",
			"test: ",
			false,
		},
		{
			"test",
			"test: 100", //int
			false,
		},
		{
			"test",
			"test: -100", //int
			true,
		},
		{
			"test",
			"test: 9223372036854775807", //int
			false,
		},
		{
			"test",
			"test: 9223372036854775808", //uintt64
			false,
		},
		{
			"test",
			"test: 19223372036854775808", //float64
			true,
		},
	}

	for i, test := range tests {
		_ = ioutil.WriteFile("current.yaml", []byte(test.entry), 0666)
		defer os.Remove("current.yaml")

		testFlag := []cli.Flag{
			&altsrc.StringFlag{StringFlag: &cli.StringFlag{Name: "conf"}},
			&altsrc.UintFlag{UintFlag: &cli.UintFlag{Name: test.name}},
		}
		app := &cli.App{}
		app.Flags = append(app.Flags, testFlag...)
		app.Before = altsrc.InitInputSourceWithContext(testFlag, altsrc.NewYamlSourceFromFlagFunc("conf"))

		appCmd := []string{"testApp", "--conf", "current.yaml"}
		err := app.Run(appCmd)
		if result := err != nil; result != test.err {
			t.Error(i, "testcast: expect error but", err)
		}
	}
}

func TestYamlFileInt64(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		err   bool
	}{
		{
			"top.test",
			`top: 
  test: 100`,
			false,
		},
		{
			"test",
			"test: ",
			false,
		},
		{
			"test",
			"test: 100", //int
			false,
		},
		{
			"test",
			"test: -100", //int
			true,
		},
		{
			"test",
			"test: 9223372036854775807", //int
			false,
		},
		{
			"test",
			"test: 9223372036854775808", //uintt64
			false,
		},
		{
			"test",
			"test: 19223372036854775808", //float64
			true,
		},
	}

	for i, test := range tests {
		_ = ioutil.WriteFile("current.yaml", []byte(test.entry), 0666)
		defer os.Remove("current.yaml")

		testFlag := []cli.Flag{
			&altsrc.StringFlag{StringFlag: &cli.StringFlag{Name: "conf"}},
			&altsrc.Int64Flag{Int64Flag: &cli.Int64Flag{Name: test.name}},
		}
		app := &cli.App{}
		app.Flags = append(app.Flags, testFlag...)
		app.Before = altsrc.InitInputSourceWithContext(testFlag, altsrc.NewYamlSourceFromFlagFunc("conf"))

		appCmd := []string{"testApp", "--conf", "current.yaml"}
		err := app.Run(appCmd)
		if result := err != nil; result != test.err {
			t.Error(i, "testcast: expect error but", err)
		}
	}
}
