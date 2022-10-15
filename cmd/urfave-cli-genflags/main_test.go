package main_test

import (
	"fmt"
	"reflect"
	"testing"

	main "github.com/urfave/cli/v2/cmd/urfave-cli-genflags"
)

func TestTypeName(t *testing.T) {
	for _, tc := range []struct {
		gt       string
		fc       *main.FlagTypeConfig
		expected string
	}{
		{gt: "int", fc: nil, expected: "IntFlag"},
		{gt: "int", fc: &main.FlagTypeConfig{}, expected: "IntFlag"},
		{gt: "int", fc: &main.FlagTypeConfig{TypeName: "VeryIntyFlag"}, expected: "VeryIntyFlag"},
		{gt: "[]bool", fc: nil, expected: "BoolSliceFlag"},
		{gt: "[]bool", fc: &main.FlagTypeConfig{}, expected: "BoolSliceFlag"},
		{gt: "[]bool", fc: &main.FlagTypeConfig{TypeName: "ManyTruthsFlag"}, expected: "ManyTruthsFlag"},
		{gt: "time.Rumination", fc: nil, expected: "RuminationFlag"},
		{gt: "time.Rumination", fc: &main.FlagTypeConfig{}, expected: "RuminationFlag"},
		{gt: "time.Rumination", fc: &main.FlagTypeConfig{TypeName: "PonderFlag"}, expected: "PonderFlag"},
	} {
		t.Run(
			fmt.Sprintf("type=%s,cfg=%v", tc.gt, func() string {
				if tc.fc != nil {
					return tc.fc.TypeName
				}
				return "nil"
			}()),
			func(ct *testing.T) {
				actual := main.TypeName(tc.gt, tc.fc)
				if tc.expected != actual {
					ct.Errorf("expected %q, got %q", tc.expected, actual)
				}
			},
		)
	}
}

func TestSpec_SortedFlagTypes(t *testing.T) {
	spec := &main.Spec{
		FlagTypes: map[string]*main.FlagTypeConfig{
			"nerf": &main.FlagTypeConfig{},
			"gerf": nil,
		},
	}

	actual := spec.SortedFlagTypes()
	expected := []*main.FlagType{
		{
			GoType: "gerf",
			Config: nil,
		},
		{
			GoType: "nerf",
			Config: &main.FlagTypeConfig{},
		},
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
}

func genFlagType() *main.FlagType {
	return &main.FlagType{
		GoType: "blerf",
		Config: &main.FlagTypeConfig{
			SkipInterfaces: []string{"fmt.Stringer"},
			StructFields: []*main.FlagStructField{
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
