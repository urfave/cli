package inputfilesupport

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
)

// LoadFlag is a default load flag used to get a inputFilePath for a yaml file
var LoadFlag = cli.StringFlag{
	Name:  "load",
	Usage: "file path to a yaml file",
}

// InitializeYaml is used to initialize Before funcs for commands.
func InitializeYaml(filePathFlagName string, flags []cli.Flag) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		filePath := context.String(filePathFlagName)
		ymlLoader := &YamlSourceLoader{FilePath: filePath}
		yamlInputSource, err := ymlLoader.Load()
		if err != nil {
			return fmt.Errorf("Unable to load Yaml file '%s': inner error: \n'%v'", filePath, err.Error())
		}

		for _, f := range flags {
			inputSourceExtendedFlag, isType := f.(FlagInputSourceExtension)
			if isType {
				inputSourceExtendedFlag.ApplyInputSourceValue(context, yamlInputSource)
			}
		}

		return nil
	}
}

// YamlSourceLoader can load yaml files and return a InputSourceContext
// to be used for a parameter value
type YamlSourceLoader struct {
	FilePath string
}

// Load returns an input source if successful or an error if there is a failure
// loading the yaml file
func (ysl *YamlSourceLoader) Load() (InputSourceContext, error) {
	var results map[string]interface{}
	err := readCommandYaml(ysl.FilePath, &results)
	if err != nil {
		return nil, err
	}

	return &MapInputSource{valueMap: results}, nil
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
	} else {
		return nil, fmt.Errorf("unable to determine how to load from path %s", filePath)
	}
}
