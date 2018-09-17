// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(http.Dir("templates"), vfsgen.Options{
		PackageName: "cli",
		VariableName: "templates",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
