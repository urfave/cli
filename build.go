//+build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli"
)

var packages = []string{"cli", "altsrc"}

func main() {
	app := cli.NewApp()

	app.Name = "builder"
	app.Usage = "Generates a new urfave/cli build!"

	app.Commands = cli.Commands{
		cli.Command{
			Name:   "vet",
			Action: VetActionFunc,
		},
		cli.Command{
			Name:   "test",
			Action: TestActionFunc,
		},
		cli.Command{
			Name:   "gfmrun",
			Action: GfmrunActionFunc,
		},
		cli.Command{
			Name:   "toc",
			Action: TocActionFunc,
		},
		cli.Command{
			Name:   "generate",
			Action: GenActionFunc,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func VetActionFunc(_ *cli.Context) error {
	cmd := exec.Command("go", "vet")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func TestActionFunc(c *cli.Context) error {
	for _, pkg := range packages {
		var packageName string

		if pkg == "cli" {
			packageName = "github.com/urfave/cli"
		} else {
			packageName = fmt.Sprintf("github.com/urfave/cli/%s", pkg)
		}

		coverProfile := fmt.Sprintf("--coverprofile=%s.coverprofile", pkg)

		cmd := exec.Command("go", "test", "-v", coverProfile, packageName)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return testCleanup()
}

func testCleanup() error {
	var out bytes.Buffer

	for _, pkg := range packages {
		file, err := os.Open(fmt.Sprintf("%s.coverprofile", pkg))
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}

		out.Write(b)
		err = file.Close()
		if err != nil {
			return err
		}

		err = os.Remove(fmt.Sprintf("%s.coverprofile", pkg))
		if err != nil {
			return err
		}
	}

	outFile, err := os.Create("coverage.txt")
	if err != nil {
		return err
	}

	_, err = out.WriteTo(outFile)
	if err != nil {
		return err
	}

	err = outFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func GfmrunActionFunc(_ *cli.Context) error {
	file, err := os.Open("README.md")
	if err != nil {
		return err
	}

	var counter int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "package main") {
			counter++
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	cmd := exec.Command("gfmrun", "-c", fmt.Sprint(counter), "-s", "README.md")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func TocActionFunc(_ *cli.Context) error {
	cmd := exec.Command("node_modules/.bin/markdown-toc", "-i", "README.md")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "diff", "--exit-code")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func GenActionFunc(_ *cli.Context) error {
	cmd := exec.Command("go", "generate", "flag-gen/main.go")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("go", "generate", "cli.go")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "diff", "--exit-code")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
