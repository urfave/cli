// +build ignore

package main

import (
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
	"log"
	"net/http"
)

func main() {
	fs := union.New(map[string]http.FileSystem{
		"/templates": http.Dir("templates"),
		"/source":    http.Dir("source"),
	})

	err := vfsgen.Generate(fs, vfsgen.Options{})

	if err != nil {
		log.Fatal(err)
	}
}
