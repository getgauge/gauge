package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

//go:generate go run genConstants.go

func main() {
	filesToRead := map[string][]string{
		"defaultProperties": {"skel", "default.properties"},
		"exampleSpec":       {"skel", "example.spec"},
		"notice":            {"notice.md"},
		"gaugeProperties":   {"skel", "gauge.properties"},
	}
	outF := filepath.Join("..", "..", "skel", "skel.go")
	out, err := os.Create(outF)
	if err != nil {
		log.Fatalf("Error creating %s\n", outF)
	}
	defer out.Close()
	out.WriteString("package skel\n\n")
	for k, v := range filesToRead {
		fp := filepath.Join(append([]string{"..", ".."}, v...)...)
		c, err := ioutil.ReadFile(fp)
		if err != nil {
			log.Fatalf("Error reading file %s\n", fp)
		}
		out.Write(append(append([]byte(fmt.Sprintf("var %s = `", k)), c...), []byte("`\n")...))
	}
}
