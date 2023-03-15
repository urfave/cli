package altsrc

import (
	"fmt"
	"path/filepath"

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
		detectableSources := cCtx.App.GetDetectableSources()

		if fileFullPath := cCtx.String(flagFileName); fileFullPath != "" {
			fileExt := filepath.Ext(fileFullPath)

			var typeError error = nil

			// Check if the App contains a handler for this extension first, allowing it to override the defaults
			if handler, ok := detectableSources[fileExt]; ok {
				source, err := handler(flagFileName)(cCtx)
				if err != nil {
					return nil, err
				}

				switch source := source.(type) {
				case InputSourceContext:
					return source, nil
				default:
					typeError = fmt.Errorf("Unable to parse config file. The type handler for %s is incorrectly implemented.", fileExt)
				}
			}

			// Fall back to the default sources implemented by the library itself
			if handler, ok := defaultSources[fileExt]; ok {
				return handler(flagFileName)(cCtx)
			}

			if typeError != nil {
				return nil, typeError
			}

			return nil, fmt.Errorf("Unable to determine config file type from extension.\nMust be one of %s", detectableExtensions(detectableSources))
		}

		return defaultInputSource()
	}
}

func detectableExtensions(detectableSources map[string]func(string) func(*cli.Context) (interface{}, error)) []string {
	detectLen := len(detectableSources)
	defaultLen := len(defaultSources)

	largerLen := defaultLen
	if detectLen > defaultLen {
		largerLen = detectLen
	}

	// The App might override some file extensions, so set size to fit the larger of the two lists, with capacity for both in case it's needed
	extensions := make([]string, largerLen, detectLen+defaultLen)

	for ext := range detectableSources {
		extensions = append(extensions, ext)
	}
	for ext := range defaultSources {
		// Only add sources that haven't been overridden by the App
		if _, ok := detectableSources[ext]; !ok {
			extensions = append(extensions, ext)
		}
	}

	return extensions
}
