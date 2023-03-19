package altsrc

import (
	"io"
	"testing"

	"github.com/urfave/cli/v2"
)

func newTestApp() *cli.App {
	a := cli.NewApp()
	a.Writer = io.Discard
	a.AddExtension(NewDetectableSourcesAppExtension())
	return a
}

func TestRegisterDetectableSource(t *testing.T) {
	app := newTestApp()
	testHandler := func(s string) func(*cli.Context) (InputSourceContext, error) {
		return func(ctx *cli.Context) (InputSourceContext, error) {
			return testInputSource{}, nil
		}
	}

	app.GetExtension("DetectableSources").(DetectableSourcesAppExtension).RegisterDetectableSource(".test", testHandler)

	_, ok := app.GetExtension("DetectableSources").(DetectableSourcesAppExtension).detectableSources[".test"]
	expect(t, ok, true)
}

func TestGetDetectableSources(t *testing.T) {
	app := newTestApp()
	testHandler := func(s string) func(*cli.Context) (InputSourceContext, error) {
		return func(ctx *cli.Context) (InputSourceContext, error) {
			return testInputSource{}, nil
		}
	}

	_, ok := app.GetExtension("DetectableSources").(DetectableSourcesAppExtension).getDetectableSources()[".test"]
	expect(t, ok, false)

	app.GetExtension("DetectableSources").(DetectableSourcesAppExtension).RegisterDetectableSource(".test", testHandler)

	_, ok = app.GetExtension("DetectableSources").(DetectableSourcesAppExtension).detectableSources[".test"]
	expect(t, ok, true)

	_, ok = app.GetExtension("DetectableSources").(DetectableSourcesAppExtension).getDetectableSources()[".test"]
	expect(t, ok, true)
}
