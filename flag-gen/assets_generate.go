// +build ignore

package main

import (
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
	"log"
	"net/http"
	"os"
	"time"
)

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
	fs := zeroModTimeFileSystem{
		Source: union.New(map[string]http.FileSystem{
			"/templates": http.Dir("templates"),
			"/source":    http.Dir("source"),
		}),
	}

	err := vfsgen.Generate(fs, vfsgen.Options{})

	if err != nil {
		log.Fatal(err)
	}
}
