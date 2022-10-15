package altsrc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"gopkg.in/yaml.v3"
)

type yamlSourceContext struct {
	FilePath string
}

// NewYamlSourceFromFile creates a new Yaml InputSourceContext from a filepath.
func NewYamlSourceFromFile(file string) (InputSourceContext, error) {
	ysc := &yamlSourceContext{FilePath: file}
	var results map[interface{}]interface{}
	err := readCommandYaml(ysc.FilePath, &results)
	if err != nil {
		return nil, fmt.Errorf("Unable to load Yaml file '%s': inner error: \n'%v'", ysc.FilePath, err.Error())
	}

	return &MapInputSource{file: file, valueMap: results}, nil
}

// NewYamlSourceFromFlagFunc creates a new Yaml InputSourceContext from a provided flag name and source context.
func NewYamlSourceFromFlagFunc(flagFileName string) func(cCtx *cli.Context) (InputSourceContext, error) {
	return func(cCtx *cli.Context) (InputSourceContext, error) {
		if filePath := cCtx.String(flagFileName); filePath != "" {
			return NewYamlSourceFromFile(filePath)
		}
		return defaultInputSource()
	}
}

func readCommandYaml(filePath string, container interface{}) (err error) {
	b, err := loadDataFrom(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, container)
	if err != nil {
		return err
	}

	err = nil
	return
}

func loadDataFrom(filePath string) ([]byte, error) {
	u, err := url.Parse(filePath)
	if err != nil {
		return nil, err
	}

	if u.Host != "" { // i have a host, now do i support the scheme?
		switch u.Scheme {
		case "http", "https":
			res, err := http.Get(filePath)
			if err != nil {
				return nil, err
			}
			return ioutil.ReadAll(res.Body)
		default:
			return nil, fmt.Errorf("scheme of %s is unsupported", filePath)
		}
	} else if u.Path != "" { // i dont have a host, but I have a path. I am a local file.
		if _, notFoundFileErr := os.Stat(filePath); notFoundFileErr != nil {
			return nil, fmt.Errorf("Cannot read from file: '%s' because it does not exist.", filePath)
		}
		return ioutil.ReadFile(filePath)
	} else if runtime.GOOS == "windows" && strings.Contains(u.String(), "\\") {
		// on Windows systems u.Path is always empty, so we need to check the string directly.
		if _, notFoundFileErr := os.Stat(filePath); notFoundFileErr != nil {
			return nil, fmt.Errorf("Cannot read from file: '%s' because it does not exist.", filePath)
		}
		return ioutil.ReadFile(filePath)
	}

	return nil, fmt.Errorf("unable to determine how to load from path %s", filePath)
}
