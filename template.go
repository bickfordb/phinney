package phinney

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
)

func (req *Request) Template(path string) (err error) {
	buf := &bytes.Buffer{}
	tmpl, err := req.App.GetTemplate(path)
	if err != nil {
		return
	}
	err = tmpl.Execute(buf, req.Context)
	if err != nil {
		return
	}
	req.Response.Header().Add("Content-Type", "text/html; charset=utf-8")
	req.Response.WriteHeader(200)
	w := bytes.NewBuffer(buf.Bytes())
	_, err = w.WriteTo(req.Response)
	return
}

type TemplateError struct {
	Path     string
	SubError error
}

func (t *TemplateError) Error() string {
	return t.SubError.Error()
}

func (app *App) getTemplate0(root *template.Template, path string, searchPaths []string) (tmpl *template.Template, err error) {
	var tmplSrcBytes []byte
	for _, searchPath := range searchPaths {
		tmplSrcBytes = app.Resources.Get(filepath.Join(searchPath, path))
		if tmplSrcBytes != nil {
			break
		}
	}
	if tmplSrcBytes == nil {
		err = &TemplateError{path, fmt.Errorf("%q does not exist", path)}
		return
	}
	tmplSrc := string(tmplSrcBytes)
	if root == nil {
		tmpl = template.New(path)
		root = tmpl
	} else {
		tmpl = root.New(path)
	}
	tmpl.Delims("{%", "%}")
	tmpl.Funcs(app.TemplateFuncs)
	tmpl, err = tmpl.Parse(tmplSrc)
	includes := parseIncludes(tmplSrc)
	for _, include := range includes {
		t := root.Lookup(include)
		if t != nil {
			continue
		}
		_, err = app.getTemplate0(root, include, searchPaths)
		if err != nil {
			return
		}
	}
	tmpl = root
	return
}

func (app *App) GetTemplate(path string) (result *template.Template, err error) {
	result, err = app.getTemplate0(nil, path, []string{".", "templates"})
	return
}

func parseIncludes(src string) (result []string) {
	pat := regexp.MustCompile(`\{\{\s*template\s+"(.+)".*\}\}`)
	result = make([]string, 0)
	for _, match := range pat.FindAllStringSubmatch(src, -1) {
		result = append(result, match[1])
	}
	return
}
