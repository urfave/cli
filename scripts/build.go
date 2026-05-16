// local build script file, similar to a makefile or collection of bash scripts in other projects

package main

import (
	"bufio"
	"bytes"
	"context"
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
	"time"

	"github.com/urfave/cli/v3"
)

const (
	badNewsEmoji      = "🚨"
	goodNewsEmoji     = "✨"
	checksPassedEmoji = "✅"

	gfmrunVersion = "v1.3.0"

	v3diffWarning = `
# The unified diff above indicates that the public API surface area
# has changed. If you feel that the changes are acceptable for the
# v3.x series, please run the following command to promote the
# current go docs:
#
#     make v3approve
#
`
)

func main() {
	topDir, err := func() (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if v, err := sh(ctx, "git", "rev-parse", "--show-toplevel"); err == nil {
			return strings.TrimSpace(v), nil
		}

		return os.Getwd()
	}()
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.Command{
		Name:  "builder",
		Usage: "Do a thing for urfave/cli! (maybe build?)",
		Commands: []*cli.Command{
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
						Sources:  cli.EnvVars("MKDOCS_REMOTE_GITHUB_TOKEN"),
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
				Name: "v3diff",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "color", Value: false},
				},
				Action: V3Diff,
			},
			{
				Name: "v3approve",
				Action: topRunAction(
					"cp",
					"-v",
					"godoc-current.txt",
					filepath.Join("testdata", "godoc-v3.x.txt"),
				),
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "tags",
				Usage: "set build tags",
			},
			&cli.StringFlag{
				Name:  "top-dir",
				Value: topDir,
			},
			&cli.StringSliceFlag{
				Name:  "packages",
				Value: []string{"cli", "scripts"},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func sh(ctx context.Context, exe string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	fmt.Fprintf(os.Stderr, "# ---> %s\n", cmd)
	outBytes, err := cmd.Output()
	return string(outBytes), err
}

func topRunAction(arg string, args ...string) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		if err := os.Chdir(cmd.String("top-dir")); err != nil {
			return err
		}

		return runCmd(ctx, arg, args...)
	}
}

func runCmd(ctx context.Context, arg string, args ...string) error {
	cmd := exec.CommandContext(ctx, arg, args...)

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
		return fmt.Errorf("download file from %[2]s into %[3]s: response %[1]v", resp.StatusCode, src, dest)
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

func VetActionFunc(ctx context.Context, cmd *cli.Command) error {
	return runCmd(ctx, "go", "vet", cmd.String("top-dir")+"/...")
}

func TestActionFunc(ctx context.Context, cmd *cli.Command) error {
	tags := cmd.String("tags")

	for _, pkg := range cmd.StringSlice("packages") {
		packageName := "github.com/urfave/cli/v3"

		if pkg != "cli" {
			packageName = fmt.Sprintf("github.com/urfave/cli/v3/%s", pkg)
		}

		args := []string{"test"}
		if tags != "" {
			args = append(args, []string{"-tags", tags}...)
		}

		args = append(args, []string{
			"-v",
			"-race",
			"--coverprofile", pkg + ".coverprofile",
			"--covermode", "atomic",
			"--cover", packageName,
			packageName,
		}...)

		if err := runCmd(ctx, "go", args...); err != nil {
			return err
		}
	}

	return testCleanup(cmd.StringSlice("packages"))
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

		fmt.Fprint(out, strings.Join(lines[1:], "\n"))

		if err := os.Remove(filename); err != nil {
			return err
		}
	}

	return os.WriteFile("coverage.txt", out.Bytes(), 0o644)
}

func GfmrunActionFunc(ctx context.Context, cmd *cli.Command) error {
	docsDir := filepath.Join(cmd.String("top-dir"), "docs")

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

	fmt.Fprintf(cmd.ErrWriter, "# ---> workspace/TMPDIR is %q\n", tmpDir)

	if err := runCmd(ctx, "go", "work", "init", docsDir); err != nil {
		return err
	}

	os.Setenv("TMPDIR", tmpDir)

	if err := os.Chdir(wd); err != nil {
		return err
	}

	dirPath := cmd.Args().Get(0)
	if dirPath == "" {
		dirPath = "README.md"
	}

	walk := cmd.Bool("walk")
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

	if err := runCmd(ctx, "gfmrun", gfmArgs...); err != nil {
		return err
	}

	return os.RemoveAll(tmpDir)
}

// checkBinarySizeActionFunc checks the size of an example binary to ensure that we are keeping size down
// this was originally inspired by https://github.com/urfave/cli/issues/1055, and followed up on as a part
// of https://github.com/urfave/cli/issues/1057
func checkBinarySizeActionFunc(ctx context.Context, cmd *cli.Command) (err error) {
	const (
		cliSourceFilePath    = "./examples/example-cli/example-cli.go"
		cliBuiltFilePath     = "./examples/example-cli/built-example"
		helloSourceFilePath  = "./examples/example-hello-world/example-hello-world.go"
		helloBuiltFilePath   = "./examples/example-hello-world/built-example"
		desiredMaxBinarySize = 2.2
		desiredMinBinarySize = 1.49
		mbStringFormatter    = "%.1fMB"
	)

	tags := cmd.String("tags")

	// get cli example size
	cliSize, err := getSize(ctx, cliSourceFilePath, cliBuiltFilePath, tags)
	if err != nil {
		return err
	}

	// get hello world size
	helloSize, err := getSize(ctx, helloSourceFilePath, helloBuiltFilePath, tags)
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
	fmt.Printf("\n%s is the current binary size\n", roundedFileSizeString)
	// show guidance for min size
	if isLessThanDesiredMin {
		fmt.Printf("  %s %s is the target min size\n", goodNewsEmoji, desiredMinSizeString)
		fmt.Println("") // visual spacing
		fmt.Println("     The binary is smaller than the target min size, which is great news!")
		fmt.Println("     That means that your changes are shrinking the binary size.")
		fmt.Println("     You'll want to go into ./scripts/build.go and decrease")
		fmt.Println("     the desiredMinBinarySize, and also probably decrease the ")
		fmt.Println("     desiredMaxBinarySize by the same amount. That will ensure that")
		fmt.Println("     future PRs will enforce the newly shrunk binary sizes.")
		fmt.Println("") // visual spacing
		os.Exit(1)
	} else {
		fmt.Printf("  %s %s is the target min size\n", checksPassedEmoji, desiredMinSizeString)
	}
	// show guidance for max size
	if isMoreThanDesiredMax {
		fmt.Printf("  %s %s is the target max size\n", badNewsEmoji, desiredMaxSizeString)
		fmt.Println("") // visual spacing
		fmt.Println("     The binary is larger than the target max size.")
		fmt.Println("     That means that your changes are increasing the binary size.")
		fmt.Println("     The first thing you'll want to do is ask your yourself")
		fmt.Println("     Is this change worth increasing the binary size?")
		fmt.Println("     Larger binary sizes for this package can dissuade its use.")
		fmt.Println("     If this change is worth the increase, then we can up the")
		fmt.Println("     desired max binary size. To do that you'll want to go into")
		fmt.Println("     ./scripts/build.go and increase the desiredMaxBinarySize,")
		fmt.Println("     and increase the desiredMinBinarySize by the same amount.")
		fmt.Println("") // visual spacing
		os.Exit(1)
	} else {
		fmt.Printf("  %s %s is the target max size\n", checksPassedEmoji, desiredMaxSizeString)
	}

	return nil
}

func GenerateActionFunc(ctx context.Context, cmd *cli.Command) error {
	topDir := cmd.String("top-dir")

	cliDocs, err := sh(ctx, "go", "doc", "-all", topDir)
	if err != nil {
		return err
	}

	return os.WriteFile(
		filepath.Join(topDir, "godoc-current.txt"),
		[]byte(cliDocs),
		0o644,
	)
}

func DiffCheckActionFunc(ctx context.Context, cmd *cli.Command) error {
	if err := os.Chdir(cmd.String("top-dir")); err != nil {
		return err
	}

	if err := runCmd(ctx, "git", "diff", "--exit-code"); err != nil {
		return err
	}

	return runCmd(ctx, "git", "diff", "--cached", "--exit-code")
}

func EnsureGoimportsActionFunc(ctx context.Context, cmd *cli.Command) error {
	topDir := cmd.String("top-dir")
	if err := os.Chdir(topDir); err != nil {
		return err
	}

	if err := runCmd(
		ctx,
		"goimports",
		"-d",
		filepath.Join(topDir, "scripts/build.go"),
	); err == nil {
		return nil
	}

	os.Setenv("GOBIN", filepath.Join(topDir, ".local/bin"))

	return runCmd(ctx, "go", "install", "golang.org/x/tools/cmd/goimports@latest")
}

func EnsureGfmrunActionFunc(ctx context.Context, cmd *cli.Command) error {
	topDir := cmd.String("top-dir")
	gfmrunExe := filepath.Join(topDir, ".local/bin/gfmrun")

	if err := os.Chdir(topDir); err != nil {
		return err
	}

	if v, err := sh(ctx, gfmrunExe, "--version"); err == nil && strings.TrimSpace(v) == gfmrunVersion {
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

	return downloadFile(gfmrunURL.String(), gfmrunExe, 0o755, 0o755)
}

func EnsureMkdocsActionFunc(ctx context.Context, cmd *cli.Command) error {
	if err := os.Chdir(cmd.String("top-dir")); err != nil {
		return err
	}

	if err := runCmd(ctx, "mkdocs", "--version"); err == nil {
		return nil
	}

	if cmd.Bool("upgrade-pip") {
		if err := runCmd(ctx, "pip", "install", "-U", "pip"); err != nil {
			return err
		}
	}

	return runCmd(ctx, "pip", "install", "-r", "mkdocs-requirements.txt")
}

func SetMkdocsRemoteActionFunc(ctx context.Context, cmd *cli.Command) error {
	ghToken := strings.TrimSpace(cmd.String("github-token"))
	if ghToken == "" {
		return errors.New("empty github token")
	}

	if err := os.Chdir(cmd.String("top-dir")); err != nil {
		return err
	}

	if err := runCmd(ctx, "git", "remote", "rm", "origin"); err != nil {
		return err
	}

	return runCmd(
		ctx,
		"git", "remote", "add", "origin",
		fmt.Sprintf("https://x-access-token:%[1]s@github.com/urfave/cli.git", ghToken),
	)
}

func LintActionFunc(ctx context.Context, cmd *cli.Command) error {
	topDir := cmd.String("top-dir")
	if err := os.Chdir(topDir); err != nil {
		return err
	}

	out, err := sh(ctx, filepath.Join(topDir, ".local/bin/goimports"), "-l", ".")
	if err != nil {
		return err
	}

	if strings.TrimSpace(out) != "" {
		fmt.Fprintln(cmd.ErrWriter, "# ---> goimports -l is non-empty:")
		fmt.Fprintln(cmd.ErrWriter, out)

		return errors.New("goimports needed")
	}

	return nil
}

func V3Diff(ctx context.Context, cmd *cli.Command) error {
	if err := os.Chdir(cmd.String("top-dir")); err != nil {
		return err
	}

	err := runCmd(
		ctx,
		"diff",
		"--ignore-all-space",
		"--minimal",
		"--color="+func() string {
			if cmd.Bool("color") {
				return "always"
			}
			return "auto"
		}(),
		"--unified",
		"--label=a/godoc",
		filepath.Join("testdata", "godoc-v3.x.txt"),
		"--label=b/godoc",
		"godoc-current.txt",
	)
	if err != nil {
		fmt.Printf("# %v ---> Hey! <---\n", badNewsEmoji)
		fmt.Println(strings.TrimSpace(v3diffWarning))
	}

	return err
}

func getSize(ctx context.Context, sourcePath, builtPath, tags string) (int64, error) {
	args := []string{"build"}

	if tags != "" {
		args = append(args, []string{"-tags", tags}...)
	}

	args = append(args, []string{
		"-o", builtPath,
		"-ldflags", "-s -w",
		sourcePath,
	}...)

	if err := runCmd(ctx, "go", args...); err != nil {
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
