package main

import (
	"encoding/json"
	"fmt"
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
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
}

// zeroModTimeFileSystem is an http.FileSystem wrapper.
// It exposes a filesystem exactly like Source, except
// all file modification times are changed to zero.
// See https://github.com/shurcooL/vfsgen/pull/40#issuecomment-355416103
type zeroModTimeFileSystem struct {
	Source http.FileSystem
}

func (fs zeroModTimeFileSystem) Open(name string) (http.File, error) {
	f, err := fs.Source.Open(name)
	return file{f}, err
}

type file struct {
	http.File
}

func (f file) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	return fileInfo{fi}, err
}

type fileInfo struct {
	os.FileInfo
}

func (fi fileInfo) ModTime() time.Time { return time.Time{} }

func main() {
	app := cli.NewApp()

	app.Name = "fg"
	app.Usage = "Generate flag type code!"
	app.Version = "v0.1.0"

	app.Action = ActionFunc

	err := GenerateAssets()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func GenerateAssets() error {
	fs := zeroModTimeFileSystem{
		Source: union.New(map[string]http.FileSystem{
			"/templates": http.Dir("templates"),
			"/source":    http.Dir("source"),
		}),
	}

	return vfsgen.Generate(fs, vfsgen.Options{
		PackageName:  "main",
		VariableName: "fs",
	})
}

func ActionFunc(_ *cli.Context) error {
	var info CliFlagInfo
	var tpl *template.Template

	inFile, err := fs.Open("/source/flag-types.json")
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
	templateFile, err := fs.Open(fmt.Sprintf("/templates/%s_flags_generated.gotpl", packageName))
	if err != nil {
		return nil, err
	}

	templateFileBytes, err := ioutil.ReadAll(templateFile)
	if err != nil {
		return nil, err
	}

	return templateFileBytes, nil
}
