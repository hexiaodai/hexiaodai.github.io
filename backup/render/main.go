package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"render/model"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

// var blogs map[string]interface{}

var blogs = &model.Blogs{}

const (
	templateFile = "template"
	blogsFile    = "blogs.yaml"
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

	if err := renderIndex(tmpl); err != nil {
		panic(err)
	}
	if err := renderTags(tmpl); err != nil {
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
		"description": func(filename string) (out string) {
			bys, err := os.ReadFile(assetsDir(filename))
			if err != nil {
				return
			}
			text := string(bys)

			lines := strings.Split(text, "\n")
			for idx, line := range lines {
				if strings.Contains(line, "# ") {
					for i := idx + 1; i < len(lines); i++ {
						if len(lines[i]) > 0 {
							out = lines[i]
							return
						}
					}
					break
				}
			}
			return
		},
	}
}

func assetsDir(filename string) string {
	return fmt.Sprintf("../%v", filename)
}

func renderIndex(tmpl *template.Template) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, blogs); err != nil {
		panic(err)
	}

	f, _ := os.Create(assetsDir("index.html"))
	defer f.Close()
	_, err := buf.WriteTo(f)
	return err
}

func renderTags(tmpl *template.Template) error {
	for _, tag := range blogs.Tags {
		newBlogs := filterBlogs(tag.Title)
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, newBlogs); err != nil {
			panic(err)
		}
		f, _ := os.Create(assetsDir(fmt.Sprintf("assets/%v.html", tag.Title)))
		if _, err := buf.WriteTo(f); err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

func filterBlogs(tag string) *model.Blogs {
	newBlogs := &model.Blogs{Tags: blogs.Tags}
	for _, blog := range blogs.Blogs {
		for _, s := range blog.Selector {
			if s == tag {
				newBlogs.Blogs = append(newBlogs.Blogs, blog)
				break
			}
		}
	}
	return newBlogs
}
