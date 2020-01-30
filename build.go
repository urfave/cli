//+build ignore

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

const packageName = "github.com/urfave/cli/v2"

func main() {
	app := cli.NewApp()

	app.Name = "builder"
	app.Usage = "Generates a new urfave/cli build!"

	app.Commands = cli.Commands{
		{
			Name:   "vet",
			Action: VetActionFunc,
		},
		{
			Name:   "test",
			Action: TestActionFunc,
		},
		{
			Name:   "gfmrun",
			Action: GfmrunActionFunc,
		},
		{
			Name:   "toc",
			Action: TocActionFunc,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runCmd(arg string, args ...string) error {
	cmd := exec.Command(arg, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func VetActionFunc(_ *cli.Context) error {
	return runCmd("go", "vet")
}

func TestActionFunc(c *cli.Context) error {
	coverProfile := fmt.Sprintf("--coverprofile=coverage.txt")

	err := runCmd("go", "test", "-v", coverProfile, packageName)
	if err != nil {
		return err
	}

	return nil
}

func GfmrunActionFunc(c *cli.Context) error {
	filename := c.Args().Get(0)
	if filename == "" {
		filename = "README.md"
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var counter int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "package main") {
			counter++
		}
	}

	err = file.Close()
	if err != nil {
		return err
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	return runCmd("gfmrun", "-c", fmt.Sprint(counter), "-s", filename)
}

func TocActionFunc(c *cli.Context) error {
	filename := c.Args().Get(0)
	if filename == "" {
		filename = "README.md"
	}

	err := runCmd("markdown-toc", "-i", filename)
	if err != nil {
		return err
	}

	err = runCmd("git", "diff", "--exit-code")
	if err != nil {
		return err
	}

	return nil
}
