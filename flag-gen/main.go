package main

//go:generate go run assets_generate.go

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

type CliFlagInfo struct {
	PackageName string
	Flags       []FlagType
}

type FlagType struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Value          bool   `json:"value"`
	Destination    bool   `json:"dest"`
	Doctail        string `json:"doctail"`
	ContextDefault string `json:"context_default"`
	ContextType    string `json:"context_type"`
	Parser         string `json:"parser"`
	ParserCast     string `json:"parser_cast"`
	ValueString    string `json:"valueString"`
	TakesFile      bool   `json:"takes_file"`
}

func main() {
	app := cli.NewApp()

	app.Name = "flag-generator"
	app.Usage = "Generate flag type code!"

	app.Action = ActionFunc

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func ActionFunc(_ *cli.Context) error {
	var info CliFlagInfo
	var tpl *template.Template

	inFile, err := assets.Open("/source/flag-types.json")
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(inFile)
	err = decoder.Decode(&info.Flags)
	if err != nil {
		log.Fatal(err)
	}

	err = inFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, packageName := range []string{"cli", "altsrc"} {
		info.PackageName = packageName

		bytes, err := ReadTemplate(packageName)
		if err != nil {
			log.Fatal(err)
		}

		tpl = template.Must(template.New("").Parse(string(bytes)))

		var outFile *os.File

		if packageName == "cli" {
			outFile, err = os.Create("flag_generated.go")
		} else {
			outFile, err = os.Create("altsrc/flag_generated.go")
		}
		if err != nil {
			log.Fatal(err)
		}

		err = tpl.Execute(outFile, info)
		if err != nil {
			log.Fatal(err)
		}

		err = outFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func ReadTemplate(packageName string) ([]byte, error) {
	templateFile, err := assets.Open(fmt.Sprintf("/templates/%s_flags_generated.gotpl", packageName))
	if err != nil {
		return nil, err
	}

	templateFileBytes, err := ioutil.ReadAll(templateFile)
	if err != nil {
		return nil, err
	}

	return templateFileBytes, nil
}
