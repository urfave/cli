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
	"sort"
	"strings"
	"syscall"
	"text/template"

	"github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	defaultPackageName = "cli"
)

var (
	//go:embed generated.gotmpl
	TemplateString string

	//go:embed generated_test.gotmpl
	TestTemplateString string

	titler = cases.Title(language.Und, cases.NoLower)
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
		Name:  "genflags",
		Usage: "Generate flag types for urfave/cli",
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

	spec := &Spec{}
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

	genTmpl, err := template.New("gen").Parse(TemplateString)
	if err != nil {
		return err
	}

	genTestTmpl, err := template.New("gen_test").Parse(TestTemplateString)
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

func TypeName(goType string, fc *FlagTypeConfig) string {
	if fc != nil && strings.TrimSpace(fc.TypeName) != "" {
		return strings.TrimSpace(fc.TypeName)
	}

	dotSplit := strings.Split(goType, ".")
	goType = dotSplit[len(dotSplit)-1]

	if strings.HasPrefix(goType, "[]") {
		return titler.String(strings.TrimPrefix(goType, "[]")) + "SliceFlag"
	}

	return titler.String(goType) + "Flag"
}

type Spec struct {
	FlagTypes              map[string]*FlagTypeConfig `yaml:"flag_types"`
	PackageName            string                     `yaml:"package_name"`
	TestPackageName        string                     `yaml:"test_package_name"`
	UrfaveCLINamespace     string                     `yaml:"urfave_cli_namespace"`
	UrfaveCLITestNamespace string                     `yaml:"urfave_cli_test_namespace"`
}

func (gfs *Spec) SortedFlagTypes() []*FlagType {
	typeNames := []string{}

	for name := range gfs.FlagTypes {
		if strings.HasPrefix(name, "[]") {
			name = strings.TrimPrefix(name, "[]") + "Slice"
		}

		typeNames = append(typeNames, name)
	}

	sort.Strings(typeNames)

	ret := make([]*FlagType, len(typeNames))

	for i, typeName := range typeNames {
		ret[i] = &FlagType{
			GoType: typeName,
			Config: gfs.FlagTypes[typeName],
		}
	}

	return ret
}

type FlagTypeConfig struct {
	SkipInterfaces []string           `yaml:"skip_interfaces"`
	StructFields   []*FlagStructField `yaml:"struct_fields"`
	TypeName       string             `yaml:"type_name"`
	ValuePointer   bool               `yaml:"value_pointer"`
}

type FlagStructField struct {
	Name string
	Type string
}

type FlagType struct {
	GoType string
	Config *FlagTypeConfig
}

func (ft *FlagType) StructFields() []*FlagStructField {
	if ft.Config == nil || ft.Config.StructFields == nil {
		return []*FlagStructField{}
	}

	return ft.Config.StructFields
}

func (ft *FlagType) ValuePointer() bool {
	if ft.Config == nil {
		return false
	}

	return ft.Config.ValuePointer
}

func (ft *FlagType) TypeName() string {
	return TypeName(ft.GoType, ft.Config)
}

func (ft *FlagType) GenerateFmtStringerInterface() bool {
	return ft.skipInterfaceNamed("fmt.Stringer")
}

func (ft *FlagType) GenerateFlagInterface() bool {
	return ft.skipInterfaceNamed("Flag")
}

func (ft *FlagType) GenerateRequiredFlagInterface() bool {
	return ft.skipInterfaceNamed("RequiredFlag")
}

func (ft *FlagType) GenerateVisibleFlagInterface() bool {
	return ft.skipInterfaceNamed("VisibleFlag")
}

func (ft *FlagType) skipInterfaceNamed(name string) bool {
	if ft.Config == nil {
		return true
	}

	lowName := strings.ToLower(name)

	for _, interfaceName := range ft.Config.SkipInterfaces {
		if strings.ToLower(interfaceName) == lowName {
			return false
		}
	}

	return true
}
