package altinputsource

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
)

var LoadFlag = cli.StringFlag{
	Name:  "load",
	Usage: "file path to a yaml file",
}

func UseYaml(c *cli.Command) {
	c.Flags = append(c.Flags, LoadFlag)
	c.FlagSetWrapperBuilder = YamlPipelineBuilder("load")
}

func YamlPipelineBuilder(filePathFlagName string) func(context *cli.Context, set *flag.FlagSet, flags []cli.Flag) (cli.FlagSetWrapper, error) {
	return func(context *cli.Context, set *flag.FlagSet, flags []cli.Flag) (cli.FlagSetWrapper, error) {
		actualFlagSetWrapper := cli.NewFlagSetWrapper(set)
		envManager := NewEnvVarFlagSetWrapper(actualFlagSetWrapper, flags)

		filePath := context.String(filePathFlagName)
		ymlLoader := &YamlSourceLoader{FilePath: filePath}
		yamlFlagSetManager, err := ymlLoader.Load(envManager)
		if err != nil {
			return nil, fmt.Errorf("Unable to load Yaml file '%s': inner error: \n'%v'", filePath, err.Error())
		}
		defaultFlagSetManager := NewDefaultValuesFlagSetWrapper(yamlFlagSetManager, flags)

		return defaultFlagSetManager, nil
	}
}

type YamlSourceLoader struct {
	FilePath string
}

func (ysl *YamlSourceLoader) Load(fsw cli.FlagSetWrapper) (cli.FlagSetWrapper, error) {
	var results map[string]interface{}
	err := readCommandYaml(ysl.FilePath, &results)
	if err != nil {
		return nil, err
	}

	return &MapFlagSetWrapper{wrappedFsw: fsw, valueMap: results}, nil
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
