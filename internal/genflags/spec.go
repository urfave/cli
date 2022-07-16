package genflags

import (
	"sort"
	"strings"
)

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
