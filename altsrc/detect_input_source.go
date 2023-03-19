package altsrc

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/urfave/cli/v2"
)

// defaultSources is a read-only map, making it concurrency-safe
var defaultSources = map[string]func(string) func(*cli.Context) (InputSourceContext, error){
	".conf": NewTomlSourceFromFlagFunc,
	".json": NewJSONSourceFromFlagFunc,
	".toml": NewTomlSourceFromFlagFunc,
	".yaml": NewYamlSourceFromFlagFunc,
	".yml":  NewYamlSourceFromFlagFunc,
}

// DetectNewSourceFromFlagFunc creates a new InputSourceContext from a provided flag name and source context.
func DetectNewSourceFromFlagFunc(flagFileName string) func(*cli.Context) (InputSourceContext, error) {
	return func(cCtx *cli.Context) (InputSourceContext, error) {
		if fileFullPath := cCtx.String(flagFileName); fileFullPath != "" {
			detectableSources := make(map[string]func(string) func(*cli.Context) (InputSourceContext, error))
			fileExt := filepath.Ext(fileFullPath)

			// Check if the App contains a handler for this extension first, allowing it to override the defaults
			detectExt, isType := cCtx.App.GetExtension("DetectableSources").(DetectableSourcesAppExtension)
			if isType {
				detectableSources = detectExt.getDetectableSources()

				if handler, ok := detectableSources[fileExt]; ok {
					return handler(flagFileName)(cCtx)
				}
			}

			// Fall back to the default sources implemented by the library itself
			if handler, ok := defaultSources[fileExt]; ok {
				return handler(flagFileName)(cCtx)
			}

			return nil, fmt.Errorf("Unable to determine config file type from extension.\nMust be one of %s", detectableExtensions(detectableSources))
		}

		return defaultInputSource()
	}
}

func detectableExtensions(detectableSources map[string]func(string) func(*cli.Context) (InputSourceContext, error)) []string {
	// We don't preallocate because this generates empty space in the output
	// It's less efficient, but this is for error messaging only at the moment
	var extensions []string

	for ext := range detectableSources {
		extensions = append(extensions, ext)
	}
	for ext := range defaultSources {
		// Only add sources that haven't been overridden by the App
		if _, ok := detectableSources[ext]; !ok {
			extensions = append(extensions, ext)
		}
	}

	sort.Strings(extensions)

	return extensions
}
