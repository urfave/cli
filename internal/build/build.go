// local build script file, similar to a makefile or collection of bash scripts in other projects

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	badNewsEmoji      = "🚨"
	goodNewsEmoji     = "✨"
	checksPassedEmoji = "✅"

	v2diffWarning = `
# The unified diff above indicates that the public API surface area
# has changed. If you feel that the changes are acceptable and adhere
# to the semantic versioning promise of the v2.x series described in
# docs/CONTRIBUTING.md, please run the following command to promote
# the current go docs:
#
#     make v2approve
#
`
)

func main() {
	top, err := func() (string, error) {
		if v, err := sh("git", "rev-parse", "--show-toplevel"); err == nil {
			return strings.TrimSpace(v), nil
		}

		return os.Getwd()
	}()
	if err != nil {
		log.Fatal(err)
	}

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
			Name: "gfmrun",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "walk",
					Value: false,
					Usage: "Walk the specified directory and perform validation on all markdown files",
				},
			},
			Action: GfmrunActionFunc,
		},
		{
			Name:   "check-binary-size",
			Action: checkBinarySizeActionFunc,
		},
		{
			Name:   "generate",
			Action: GenerateActionFunc,
		},
		{
			Name: "v2diff",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "color", Value: false},
			},
			Action: V2Diff,
		},
		{
			Name:   "v2approve",
			Action: V2Approve,
		},
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "tags",
			Usage: "set build tags",
		},
		&cli.PathFlag{
			Name:  "top",
			Value: top,
		},
		&cli.StringSliceFlag{
			Name:  "packages",
			Value: cli.NewStringSlice("cli", "altsrc", "internal/build"),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func sh(exe string, args ...string) (string, error) {
	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	fmt.Fprintf(os.Stderr, "# ---> %s\n", cmd)
	outBytes, err := cmd.Output()
	return string(outBytes), err
}

func runCmd(arg string, args ...string) error {
	cmd := exec.Command(arg, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Fprintf(os.Stderr, "# ---> %s\n", cmd)
	return cmd.Run()
}

func VetActionFunc(cCtx *cli.Context) error {
	return runCmd("go", "vet", cCtx.Path("top")+"/...")
}

func TestActionFunc(c *cli.Context) error {
	tags := c.String("tags")

	for _, pkg := range c.StringSlice("packages") {
		packageName := "github.com/urfave/cli/v2"

		if pkg != "cli" {
			packageName = fmt.Sprintf("github.com/urfave/cli/v2/%s", pkg)
		}

		if err := runCmd(
			"go", "test",
			"-tags", tags,
			"-v",
			"--coverprofile", pkg+".coverprofile",
			"--covermode", "count",
			"--cover", packageName,
			packageName,
		); err != nil {
			return err
		}
	}

	return testCleanup(c.StringSlice("packages"))
}

func testCleanup(packages []string) error {
	out := &bytes.Buffer{}

	fmt.Fprintf(out, "mode: count\n")

	for _, pkg := range packages {
		filename := pkg + ".coverprofile"

		lineBytes, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		lines := strings.Split(string(lineBytes), "\n")

		fmt.Fprintf(out, strings.Join(lines[1:], "\n"))

		if err := os.Remove(filename); err != nil {
			return err
		}
	}

	return os.WriteFile("coverage.txt", out.Bytes(), 0644)
}

func GfmrunActionFunc(cCtx *cli.Context) error {
	top := cCtx.Path("top")

	bash, err := exec.LookPath("bash")
	if err != nil {
		return err
	}

	os.Setenv("SHELL", bash)

	tmpDir, err := os.MkdirTemp("", "urfave-cli*")
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.Chdir(tmpDir); err != nil {
		return err
	}

	fmt.Fprintf(cCtx.App.ErrWriter, "# ---> workspace/TMPDIR is %q\n", tmpDir)

	if err := runCmd("go", "work", "init", top); err != nil {
		return err
	}

	os.Setenv("TMPDIR", tmpDir)

	if err := os.Chdir(wd); err != nil {
		return err
	}

	dirPath := cCtx.Args().Get(0)
	if dirPath == "" {
		dirPath = "README.md"
	}

	walk := cCtx.Bool("walk")
	sources := []string{}

	if walk {
		// Walk the directory and find all markdown files.
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".md" {
				return nil
			}

			sources = append(sources, path)
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		sources = append(sources, dirPath)
	}

	var counter int

	for _, src := range sources {
		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()

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
	}

	gfmArgs := []string{
		"--count",
		fmt.Sprint(counter),
	}
	for _, src := range sources {
		gfmArgs = append(gfmArgs, "--sources", src)
	}

	if err := runCmd("gfmrun", gfmArgs...); err != nil {
		return err
	}

	return os.RemoveAll(tmpDir)
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
		desiredMaxBinarySize = 2.2
		mbStringFormatter    = "%.1fMB"
	)

	desiredMinBinarySize := 1.675

	tags := c.String("tags")

	if strings.Contains(tags, "urfave_cli_no_docs") {
		desiredMinBinarySize = 1.39
	}

	// get cli example size
	cliSize, err := getSize(cliSourceFilePath, cliBuiltFilePath, tags)
	if err != nil {
		return err
	}

	// get hello world size
	helloSize, err := getSize(helloSourceFilePath, helloBuiltFilePath, tags)
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

func GenerateActionFunc(cCtx *cli.Context) error {
	top := cCtx.Path("top")

	cliDocs, err := sh("go", "doc", "-all", top)
	if err != nil {
		return err
	}

	altsrcDocs, err := sh("go", "doc", "-all", filepath.Join(top, "altsrc"))
	if err != nil {
		return err
	}

	if err := os.WriteFile(
		filepath.Join(top, "godoc-current.txt"),
		[]byte(cliDocs+altsrcDocs),
		0644,
	); err != nil {
		return err
	}

	return runCmd("go", "generate", cCtx.Path("top")+"/...")
}

func V2Diff(cCtx *cli.Context) error {
	os.Chdir(cCtx.Path("top"))

	err := runCmd(
		"diff",
		"--ignore-all-space",
		"--minimal",
		"--color="+func() string {
			if cCtx.Bool("color") {
				return "always"
			}
			return "auto"
		}(),
		"--unified",
		"--label=a/godoc",
		filepath.Join("testdata", "godoc-v2.x.txt"),
		"--label=b/godoc",
		"godoc-current.txt",
	)

	if err != nil {
		fmt.Printf("# %v ---> Hey! <---\n", badNewsEmoji)
		fmt.Println(strings.TrimSpace(v2diffWarning))
	}

	return err
}

func V2Approve(cCtx *cli.Context) error {
	top := cCtx.Path("top")

	return runCmd(
		"cp",
		"-v",
		filepath.Join(top, "godoc-current.txt"),
		filepath.Join(top, "testdata", "godoc-v2.x.txt"),
	)
}

func getSize(sourcePath string, builtPath string, tags string) (size int64, err error) {
	// build example binary
	err = runCmd("go", "build", "-tags", tags, "-o", builtPath, "-ldflags", "-s -w", sourcePath)
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
