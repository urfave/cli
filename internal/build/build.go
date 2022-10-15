// local build script file, similar to a makefile or collection of bash scripts in other projects

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	badNewsEmoji      = "🚨"
	goodNewsEmoji     = "✨"
	checksPassedEmoji = "✅"

	gfmrunVersion = "v1.3.0"

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

	app := &cli.App{
		Name:  "builder",
		Usage: "Do a thing for urfave/cli! (maybe build?)",
		Commands: cli.Commands{
			{
				Name:   "vet",
				Action: topRunAction("go", "vet", "./..."),
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
				Name: "yamlfmt",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "strict", Value: false, Usage: "require presence of yq"},
				},
				Action: YAMLFmtActionFunc,
			},
			{
				Name:   "diffcheck",
				Action: DiffCheckActionFunc,
			},
			{
				Name:   "ensure-goimports",
				Action: EnsureGoimportsActionFunc,
			},
			{
				Name:   "ensure-gfmrun",
				Action: EnsureGfmrunActionFunc,
			},
			{
				Name:   "ensure-mkdocs",
				Action: EnsureMkdocsActionFunc,
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "upgrade-pip"},
				},
			},
			{
				Name:   "set-mkdocs-remote",
				Action: SetMkdocsRemoteActionFunc,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "github-token",
						EnvVars:  []string{"MKDOCS_REMOTE_GITHUB_TOKEN"},
						Required: true,
					},
				},
			},
			{
				Name:   "deploy-mkdocs",
				Action: topRunAction("mkdocs", "gh-deploy", "--force"),
			},
			{
				Name:   "lint",
				Action: LintActionFunc,
			},
			{
				Name: "v2diff",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "color", Value: false},
				},
				Action: V2Diff,
			},
			{
				Name: "v2approve",
				Action: topRunAction(
					"cp",
					"-v",
					"godoc-current.txt",
					filepath.Join("testdata", "godoc-v2.x.txt"),
				),
			},
		},
		Flags: []cli.Flag{
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

func topRunAction(arg string, args ...string) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		os.Chdir(cCtx.Path("top"))

		return runCmd(arg, args...)
	}
}

func runCmd(arg string, args ...string) error {
	cmd := exec.Command(arg, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Fprintf(os.Stderr, "# ---> %s\n", cmd)
	return cmd.Run()
}

func downloadFile(src, dest string, dirPerm, perm os.FileMode) error {
	req, err := http.NewRequest(http.MethodGet, src, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("download response %[1]v", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dest), dirPerm); err != nil {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return os.Chmod(dest, perm)
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

		args := []string{"test"}
		if tags != "" {
			args = append(args, []string{"-tags", tags}...)
		}

		args = append(args, []string{
			"-v",
			"--coverprofile", pkg + ".coverprofile",
			"--covermode", "count",
			"--cover", packageName,
			packageName,
		}...)

		if err := runCmd("go", args...); err != nil {
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

func YAMLFmtActionFunc(cCtx *cli.Context) error {
	yqBin, err := exec.LookPath("yq")
	if err != nil {
		if !cCtx.Bool("strict") {
			fmt.Fprintln(cCtx.App.ErrWriter, "# ---> no yq found; skipping")
			return nil
		}

		return err
	}

	os.Chdir(cCtx.Path("top"))

	return runCmd(yqBin, "eval", "--inplace", "flag-spec.yaml")
}

func DiffCheckActionFunc(cCtx *cli.Context) error {
	os.Chdir(cCtx.Path("top"))

	if err := runCmd("git", "diff", "--exit-code"); err != nil {
		return err
	}

	return runCmd("git", "diff", "--cached", "--exit-code")
}

func EnsureGoimportsActionFunc(cCtx *cli.Context) error {
	top := cCtx.Path("top")
	os.Chdir(top)

	if err := runCmd(
		"goimports",
		"-d",
		filepath.Join(top, "internal/build/build.go"),
	); err == nil {
		return nil
	}

	os.Setenv("GOBIN", filepath.Join(top, ".local/bin"))

	return runCmd("go", "install", "golang.org/x/tools/cmd/goimports@latest")
}

func EnsureGfmrunActionFunc(cCtx *cli.Context) error {
	top := cCtx.Path("top")
	gfmrunExe := filepath.Join(top, ".local/bin/gfmrun")

	os.Chdir(top)

	if v, err := sh(gfmrunExe, "--version"); err == nil && strings.TrimSpace(v) == gfmrunVersion {
		return nil
	}

	gfmrunURL, err := url.Parse(
		fmt.Sprintf(
			"https://github.com/urfave/gfmrun/releases/download/%[1]s/gfmrun-%[2]s-%[3]s-%[1]s",
			gfmrunVersion, runtime.GOOS, runtime.GOARCH,
		),
	)
	if err != nil {
		return err
	}

	return downloadFile(gfmrunURL.String(), gfmrunExe, 0755, 0755)
}

func EnsureMkdocsActionFunc(cCtx *cli.Context) error {
	os.Chdir(cCtx.Path("top"))

	if err := runCmd("mkdocs", "--version"); err == nil {
		return nil
	}

	if cCtx.Bool("upgrade-pip") {
		if err := runCmd("pip", "install", "-U", "pip"); err != nil {
			return err
		}
	}

	return runCmd("pip", "install", "-r", "mkdocs-requirements.txt")
}

func SetMkdocsRemoteActionFunc(cCtx *cli.Context) error {
	ghToken := strings.TrimSpace(cCtx.String("github-token"))
	if ghToken == "" {
		return errors.New("empty github token")
	}

	os.Chdir(cCtx.Path("top"))

	if err := runCmd("git", "remote", "rm", "origin"); err != nil {
		return err
	}

	return runCmd(
		"git", "remote", "add", "origin",
		fmt.Sprintf("https://x-access-token:%[1]s@github.com/urfave/cli.git", ghToken),
	)
}

func LintActionFunc(cCtx *cli.Context) error {
	top := cCtx.Path("top")
	os.Chdir(top)

	out, err := sh(filepath.Join(top, ".local/bin/goimports"), "-l", ".")
	if err != nil {
		return err
	}

	if strings.TrimSpace(out) != "" {
		fmt.Fprintln(cCtx.App.ErrWriter, "# ---> goimports -l is non-empty:")
		fmt.Fprintln(cCtx.App.ErrWriter, out)

		return errors.New("goimports needed")
	}

	return nil
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

func getSize(sourcePath, builtPath, tags string) (int64, error) {
	args := []string{"build"}

	if tags != "" {
		args = append(args, []string{"-tags", tags}...)
	}

	args = append(args, []string{
		"-o", builtPath,
		"-ldflags", "-s -w",
		sourcePath,
	}...)

	if err := runCmd("go", args...); err != nil {
		fmt.Println("issue getting size for example binary")
		return 0, err
	}

	fileInfo, err := os.Stat(builtPath)
	if err != nil {
		fmt.Println("issue getting size for example binary")
		return 0, err
	}

	return fileInfo.Size(), nil
}
