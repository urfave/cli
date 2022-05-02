package main

import (
	"bytes"
	"context"
	_ "embed"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/internal/genflags"
	"gopkg.in/yaml.v2"
)

const (
	defaultPackageName = "cli"
)

func sh(ctx context.Context, exe string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Stderr = os.Stderr
	outBytes, err := cmd.Output()
	return string(outBytes), err
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	top := "../../"
	if v, err := sh(ctx, "git", "rev-parse", "--show-toplevel"); err == nil {
		top = strings.TrimSpace(v)
	}

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "flag-spec-yaml",
				Aliases: []string{"f"},
				Value:   filepath.Join(top, "flag-spec.yaml"),
			},
			&cli.PathFlag{
				Name:    "generated-output",
				Aliases: []string{"o"},
				Value:   filepath.Join(top, "zz_generated.flags.go"),
			},
			&cli.PathFlag{
				Name:    "generated-test-output",
				Aliases: []string{"t"},
				Value:   filepath.Join(top, "zz_generated.flags_test.go"),
			},
			&cli.StringFlag{
				Name:    "generated-package-name",
				Aliases: []string{"p"},
				Value:   defaultPackageName,
			},
			&cli.StringFlag{
				Name:    "generated-test-package-name",
				Aliases: []string{"T"},
				Value:   defaultPackageName + "_test",
			},
			&cli.StringFlag{
				Name:    "urfave-cli-namespace",
				Aliases: []string{"n"},
				Value:   "",
			},
			&cli.StringFlag{
				Name:    "urfave-cli-test-namespace",
				Aliases: []string{"N"},
				Value:   "cli.",
			},
		},
		Action: runGenFlags,
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

func runGenFlags(cCtx *cli.Context) error {
	specBytes, err := os.ReadFile(cCtx.Path("flag-spec-yaml"))
	if err != nil {
		return err
	}

	spec := &genflags.Spec{}
	if err := yaml.Unmarshal(specBytes, spec); err != nil {
		return err
	}

	if cCtx.IsSet("generated-package-name") {
		spec.PackageName = strings.TrimSpace(cCtx.String("generated-package-name"))
	}

	if strings.TrimSpace(spec.PackageName) == "" {
		spec.PackageName = defaultPackageName
	}

	if cCtx.IsSet("generated-test-package-name") {
		spec.TestPackageName = strings.TrimSpace(cCtx.String("generated-test-package-name"))
	}

	if strings.TrimSpace(spec.TestPackageName) == "" {
		spec.TestPackageName = defaultPackageName + "_test"
	}

	if cCtx.IsSet("urfave-cli-namespace") {
		spec.UrfaveCLINamespace = strings.TrimSpace(cCtx.String("urfave-cli-namespace"))
	}

	if cCtx.IsSet("urfave-cli-test-namespace") {
		spec.UrfaveCLITestNamespace = strings.TrimSpace(cCtx.String("urfave-cli-test-namespace"))
	} else {
		spec.UrfaveCLITestNamespace = "cli."
	}

	genTmpl, err := template.New("gen").Parse(genflags.TemplateString)
	if err != nil {
		return err
	}

	genTestTmpl, err := template.New("gen_test").Parse(genflags.TestTemplateString)
	if err != nil {
		return err
	}

	genBuf := &bytes.Buffer{}
	if err := genTmpl.Execute(genBuf, spec); err != nil {
		return err
	}

	genTestBuf := &bytes.Buffer{}
	if err := genTestTmpl.Execute(genTestBuf, spec); err != nil {
		return err
	}

	if err := os.WriteFile(cCtx.Path("generated-output"), genBuf.Bytes(), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(cCtx.Path("generated-test-output"), genTestBuf.Bytes(), 0644); err != nil {
		return err
	}

	if _, err := sh(cCtx.Context, "goimports", "-w", cCtx.Path("generated-output")); err != nil {
		return err
	}

	if _, err := sh(cCtx.Context, "goimports", "-w", cCtx.Path("generated-test-output")); err != nil {
		return err
	}

	return nil
}
