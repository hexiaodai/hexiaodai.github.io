package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

var blogs map[interface{}]interface{}

const (
	templateFile = "template"
	blogsFile    = "blogs.yaml"
	htmlFile     = "../index.html"
)

func main() {
	templateByte, err := os.ReadFile(templateFile)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New("index").Funcs(funcMap()).Parse(string(templateByte))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, blogs); err != nil {
		panic(err)
	}

	f, _ := os.Create(htmlFile)
	defer f.Close()
	if _, err := buf.WriteTo(f); err != nil {
		panic(err)
	}
}

func init() {
	data, err := os.ReadFile(blogsFile)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(data, &blogs); err != nil {
		panic(err)
	}
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"href": func(filename string) string {
			extension := filepath.Ext(filename)
			return fmt.Sprintf("%v.html", filename[0:len(filename)-len(extension)])
		},
	}
}
