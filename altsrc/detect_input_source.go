package altsrc

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var detectableSources = map[string]func(string) func(cCtx *cli.Context) (InputSourceContext, error){
	".conf": NewTomlSourceFromFlagFunc,
	".json": NewJSONSourceFromFlagFunc,
	".toml": NewTomlSourceFromFlagFunc,
	".yaml": NewYamlSourceFromFlagFunc,
	".yml":  NewYamlSourceFromFlagFunc,
}

// RegisterDetectableSource lets developers add support for their own altsrc filetypes to the autodetection list.
func RegisterDetectableSource(extension string, handler func(string) func(cCtx *cli.Context) (InputSourceContext, error)) {
	detectableSources[extension] = handler
}

// DetectNewSourceFromFlagFunc creates a new InputSourceContext from a provided flag name and source context.
func DetectNewSourceFromFlagFunc(flagFileName string) func(cCtx *cli.Context) (InputSourceContext, error) {
	return func(cCtx *cli.Context) (InputSourceContext, error) {
		if fileFullPath := cCtx.String(flagFileName); fileFullPath != "" {
			fileExt := filepath.Ext(fileFullPath)
			if handler, ok := detectableSources[fileExt]; ok {
				return handler(flagFileName)(cCtx)
			}
			return nil, fmt.Errorf("Unable to determine config file type from extension.\nMust be one of %v", detectableExtensions())
		}
		return defaultInputSource()
	}
}

func detectableExtensions() []string {
	extensions := make([]string, len(detectableSources))
	for ext := range detectableSources {
		extensions = append(extensions, ext)
	}

	return extensions
}
