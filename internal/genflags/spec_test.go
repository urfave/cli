package genflags_test

import (
	"reflect"
	"testing"

	"github.com/urfave/cli/v2/internal/genflags"
)

func TestSpec_SortedFlagTypes(t *testing.T) {
	spec := &genflags.Spec{
		FlagTypes: map[string]*genflags.FlagTypeConfig{
			"nerf": &genflags.FlagTypeConfig{},
			"gerf": nil,
		},
	}

	actual := spec.SortedFlagTypes()
	expected := []*genflags.FlagType{
		{
			GoType: "gerf",
			Config: nil,
		},
		{
			GoType: "nerf",
			Config: &genflags.FlagTypeConfig{},
		},
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
}

func genFlagType() *genflags.FlagType {
	return &genflags.FlagType{
		GoType: "blerf",
		Config: &genflags.FlagTypeConfig{
			SkipInterfaces: []string{"fmt.Stringer"},
			StructFields: []*genflags.FlagStructField{
				{
					Name: "Foibles",
					Type: "int",
				},
				{
					Name: "Hoopled",
					Type: "bool",
				},
			},
			TypeName:     "YeOldeBlerfFlag",
			ValuePointer: true,
		},
	}
}

func TestFlagType_StructFields(t *testing.T) {
	ft := genFlagType()

	sf := ft.StructFields()
	if 2 != len(sf) {
		t.Errorf("expected 2 struct fields, got %v", len(sf))
		return
	}

	if "Foibles" != sf[0].Name {
		t.Errorf("expected struct field order to be retained")
	}
}

func TestFlagType_ValuePointer(t *testing.T) {
	ft := genFlagType()

	if !ft.ValuePointer() {
		t.Errorf("expected ValuePointer to be true")
		return
	}

	ft.Config = nil

	if ft.ValuePointer() {
		t.Errorf("expected ValuePointer to be false")
	}
}

func TestFlagType_GenerateFmtStringerInterface(t *testing.T) {
	ft := genFlagType()

	if ft.GenerateFmtStringerInterface() {
		t.Errorf("expected GenerateFmtStringerInterface to be false")
		return
	}

	ft.Config = nil

	if !ft.GenerateFmtStringerInterface() {
		t.Errorf("expected GenerateFmtStringerInterface to be true")
	}
}

func TestFlagType_GenerateFlagInterface(t *testing.T) {
	ft := genFlagType()

	if !ft.GenerateFlagInterface() {
		t.Errorf("expected GenerateFlagInterface to be true")
		return
	}

	ft.Config = nil

	if !ft.GenerateFlagInterface() {
		t.Errorf("expected GenerateFlagInterface to be true")
	}
}
