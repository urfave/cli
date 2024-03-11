package altsrc

import (
	"github.com/urfave/cli/v2"
)

type DetectableSourcesAppExtension struct {
	detectableSources map[string]func(string) func(*cli.Context) (InputSourceContext, error)
}

func NewDetectableSourcesAppExtension() DetectableSourcesAppExtension {
	return DetectableSourcesAppExtension{
		detectableSources: make(map[string]func(string) func(*cli.Context) (InputSourceContext, error)),
	}
}

// MyName satisfies the cli.AppExtension interface, providing a name to register the extension under
func (e DetectableSourcesAppExtension) MyName() string {
	return "DetectableSources"
}

// RegisterDetectableSource lets developers add support for their own altsrc filetypes to the autodetection list.
func (e DetectableSourcesAppExtension) RegisterDetectableSource(extension string, handler func(string) func(*cli.Context) (InputSourceContext, error)) {
	if e.detectableSources == nil {
		e.detectableSources = make(map[string]func(string) func(*cli.Context) (InputSourceContext, error))
	}
	e.detectableSources[extension] = handler
}

func (e DetectableSourcesAppExtension) getDetectableSources() map[string]func(string) func(*cli.Context) (InputSourceContext, error) {
	return e.detectableSources
}
