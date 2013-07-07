package phinney

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"text/template"
  "net/url"
  "regexp"
)

func (app *App) CSSBundle(pathTemplateSrc string, routeName string, resourcePaths []string) {
	app.Bundle(pathTemplateSrc, routeName, resourcePaths, "text/css")
}

func (app *App) JSBundle(pathTemplateSrc string, routeName string, resourcePaths []string) {
	app.Bundle(pathTemplateSrc, routeName, resourcePaths, "application/javascript")
}

func (app *App) Bundle(pathTemplateSrc string, routeName string, resourcePaths []string, contentType string) {
	pathTemplate, err := template.New("path").Parse(pathTemplateSrc)
	if err != nil {
		app.Errors <- MakeAppError("bundle", err.Error())
		return
	}
	route := &Route{}
	route.Name = routeName
	updateData := func(data []byte) {
		var err error
		if err != nil {
			app.Errors <- MakeAppError("bundle", err.Error())
			return
		}
		version := fmt.Sprintf("%X", crc32.ChecksumIEEE(data))
		route.Reverse, err = renderTemplate(pathTemplate, map[string]string{"Version": version})
		if err != nil {
			app.Errors <- MakeAppError("bundle", err.Error())
			return
		}
		route.Pattern, err = urlToPathPattern(route.Reverse)
		if err != nil {
			app.Errors <- MakeAppError("bundle", err.Error())
			return
		}
		route.Handler = ServeBytes(data, map[string]string{"Content-Type": contentType})
	}
	updateData([]byte{})
	app.Router.Add(route)

	go func() {
		buf := &bytes.Buffer{}
		for _, resourcePath := range resourcePaths {
			data := app.Resources.Get(resourcePath)
			if data == nil {
				app.Errors <- MakeAppError("bundle", fmt.Sprintf("expecting %q", resourcePath))
				return
			} else {
				buf.Write(data)
			}
		}
		updateData(buf.Bytes())
	}()
}

func renderTemplate(t *template.Template, context interface{}) (result string, err error) {
	buf := &bytes.Buffer{}
	err = t.Execute(buf, context)
	if err != nil {
		return
	}
	result = string(buf.Bytes())
	return
}

func urlToPathPattern(urlS string) (result *regexp.Regexp, err error) {
	u, err := url.Parse(urlS)
	if err != nil {
		return
	}
	p := "^" + regexp.QuoteMeta(u.Path) + "$"
	result, err = regexp.Compile(p)
	return
}
