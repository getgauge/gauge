package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
)

//go:generate go run constants.go

func main() {
	createGaugeProperties()
	filesToRead := map[string][]string{
		"defaultProperties": {"skel", "default.properties"},
		"exampleSpec":       {"skel", "example.spec"},
		"notice":            {"notice.md"},
		"gaugeProperties":   {"skel", common.GaugePropertiesFile},
		"gitignore":         {"skel", ".gitignore"},
	}
	goPath := os.Getenv("GOPATH")
	outF := filepath.Join(goPath, "src", "github.com", "getgauge", "gauge", "skel", "skel.go")
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

func createGaugeProperties() {
	goPath := os.Getenv("GOPATH")
	outF := filepath.Join(goPath, "src", "github.com", "getgauge", "gauge", "skel", common.GaugePropertiesFile)
	out, err := os.Create(outF)
	if err != nil {
		log.Fatalf("Error creating %s\n", outF)
	}
	defer out.Close()
	if _, err := (config.NewProperties()).Write(out); err != nil {
		log.Fatalf("Error Writing %s\n", outF)
	}
}
