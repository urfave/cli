// local build script file, similar to a makefile or collection of bash scripts in other projects

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

var packages = []string{"cli", "altsrc"}

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
		{
			Name:   "check-binary-size",
			Action: checkBinarySizeActionFunc,
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
	for _, pkg := range packages {
		var packageName string

		if pkg == "cli" {
			packageName = "github.com/urfave/cli/v2"
		} else {
			packageName = fmt.Sprintf("github.com/urfave/cli/v2/%s", pkg)
		}

		coverProfile := fmt.Sprintf("--coverprofile=%s.coverprofile", pkg)

		err := runCmd("go", "test", "-v", coverProfile, packageName)
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

// checkBinarySizeActionFunc checks the size of an example binary to ensure that we are keeping size down
// this was originally inspired by https://github.com/urfave/cli/issues/1055, and followed up on as a part
// of https://github.com/urfave/cli/issues/1057
func checkBinarySizeActionFunc(c *cli.Context) (err error) {
	const (
		cliSourceFilePath    = "./internal/example-cli/example-cli.go"
		cliBuiltFilePath     = "./internal/example-cli/built-example"
		helloSourceFilePath  = "./internal/example-hello-world/example-hello-world.go"
		helloBuiltFilePath   = "./internal/example-hello-world/built-example"
		desiredMinBinarySize = 2.0
		desiredMaxBinarySize = 2.1
		badNewsEmoji         = "🚨"
		goodNewsEmoji        = "✨"
		checksPassedEmoji    = "✅"
		mbStringFormatter    = "%.1fMB"
	)

	// get cli example size
	cliSize, err := getSize(cliSourceFilePath, cliBuiltFilePath)
	if err != nil {
		return err
	}

	// get hello world size
	helloSize, err := getSize(helloSourceFilePath, helloBuiltFilePath)
	if err != nil {
		return err
	}

	// The CLI size diff is the number we are interested in.
	// This tells us how much our CLI package contributes to the binary size.
	cliSizeDiff := cliSize - helloSize

	// get human readable size, in MB with one decimal place.
	// example output is: 35.2MB. (note: this simply an example)
	// that output is much easier to reason about than the `35223432`
	// that you would see output without the rounding
	fileSizeInMB := float64(cliSizeDiff) / float64(1000000)
	roundedFileSize := math.Round(fileSizeInMB*10) / 10
	roundedFileSizeString := fmt.Sprintf(mbStringFormatter, roundedFileSize)

	// check against bounds
	isLessThanDesiredMin := roundedFileSize < desiredMinBinarySize
	isMoreThanDesiredMax := roundedFileSize > desiredMaxBinarySize
	desiredMinSizeString := fmt.Sprintf(mbStringFormatter, desiredMinBinarySize)
	desiredMaxSizeString := fmt.Sprintf(mbStringFormatter, desiredMaxBinarySize)

	// show guidance
	fmt.Println(fmt.Sprintf("\n%s is the current binary size", roundedFileSizeString))
	// show guidance for min size
	if isLessThanDesiredMin {
		fmt.Println(fmt.Sprintf("  %s %s is the target min size", goodNewsEmoji, desiredMinSizeString))
		fmt.Println("") // visual spacing
		fmt.Println("     The binary is smaller than the target min size, which is great news!")
		fmt.Println("     That means that your changes are shrinking the binary size.")
		fmt.Println("     You'll want to go into ./internal/build/build.go and decrease")
		fmt.Println("     the desiredMinBinarySize, and also probably decrease the ")
		fmt.Println("     desiredMaxBinarySize by the same amount. That will ensure that")
		fmt.Println("     future PRs will enforce the newly shrunk binary sizes.")
		fmt.Println("") // visual spacing
		os.Exit(1)
	} else {
		fmt.Println(fmt.Sprintf("  %s %s is the target min size", checksPassedEmoji, desiredMinSizeString))
	}
	// show guidance for max size
	if isMoreThanDesiredMax {
		fmt.Println(fmt.Sprintf("  %s %s is the target max size", badNewsEmoji, desiredMaxSizeString))
		fmt.Println("") // visual spacing
		fmt.Println("     The binary is larger than the target max size.")
		fmt.Println("     That means that your changes are increasing the binary size.")
		fmt.Println("     The first thing you'll want to do is ask your yourself")
		fmt.Println("     Is this change worth increasing the binary size?")
		fmt.Println("     Larger binary sizes for this package can dissuade its use.")
		fmt.Println("     If this change is worth the increase, then we can up the")
		fmt.Println("     desired max binary size. To do that you'll want to go into")
		fmt.Println("     ./internal/build/build.go and increase the desiredMaxBinarySize,")
		fmt.Println("     and increase the desiredMinBinarySize by the same amount.")
		fmt.Println("") // visual spacing
		os.Exit(1)
	} else {
		fmt.Println(fmt.Sprintf("  %s %s is the target max size", checksPassedEmoji, desiredMaxSizeString))
	}

	return nil
}

func getSize(sourcePath string, builtPath string) (size int64, err error) {
	// build example binary
	err = runCmd("go", "build", "-o", builtPath, "-ldflags", "-s -w", sourcePath)
	if err != nil {
		fmt.Println("issue getting size for example binary")
		return 0, err
	}

	// get file info
	fileInfo, err := os.Stat(builtPath)
	if err != nil {
		fmt.Println("issue getting size for example binary")
		return 0, err
	}

	// size!
	size = fileInfo.Size()

	return size, nil
}
