package phinney

import (
	"strings"
  "html/template"
)

type AppError interface {
	Component() string
	Error() string
}

type simpleAppError struct {
	c string
	e string
}

func (s *simpleAppError) Component() string {
	return s.c
}

func (s *simpleAppError) Error() string {
	return s.e
}

func MakeAppError(component string, msg string) AppError {
	return &simpleAppError{component, msg}
}

type App struct {
	Router    *Router
	Resources *Resources
	Errors    chan AppError
  TemplateFuncs template.FuncMap
  Plugins []*Plugin
}

func NewApp() (app *App) {
  app = &App{}
	e := envDict()
	envResourcesPath, exists := e["RESOURCES_PATH"]
	resourcePaths := make([]string, 0)
	if exists {
		for _, p := range strings.Split(envResourcesPath, ":") {
			if len(p) > 0 {
				log.Debug("resource path %q", p)
				resourcePaths = append(resourcePaths, p)
			}
		}
	}
	errors := make(chan AppError)
  app.Router = NewRouter()
  app.Resources = NewResources(resourcePaths)
  app.Errors = errors
  app.TemplateFuncs = make(template.FuncMap)
  app.TemplateFuncs["Reverse"] = app.Router.Reverse
  app.Plugins = make([]*Plugin, 0)
  app.Plugins = append(app.Plugins, NewRequestContextPlugin())
	return
}

func (app *App) Main() {
	e := <-app.Errors
	log.Error("component %q had unexpected error %q", e.Component(), e.Error())
}
