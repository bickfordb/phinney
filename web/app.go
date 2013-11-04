package web

import (
  "fmt"
  "net/http"
  "reflect"
  "strings"
)

type App struct {
  routes []route
  templateFuncs map[string]interface{}
}

func (app *App) TemplateFunc(key string, f interface{}) {
  if (app.templateFuncs == nil) {
    app.templateFuncs = make(map[string]interface{})
  }
  app.templateFuncs[key] = f
}

func (app *App) ServeHTTP(response http.ResponseWriter, request *http.Request) {
  var err error
  path := request.URL.Path
  handlerVal := &concreteHandler{}
  handlerVal.request = request
  handlerVal.response = response
  handlerVal.status = http.StatusOK
  handlerVal.app = app
  for _, route := range app.routes {
    fmt.Println("route: ", route)
    matches, args := route.matches(path)
    if !matches {
      continue
    }
    handlerVal.args = args
    println("found handler");
    val := reflect.ValueOf(route.handler)
    h := reflect.New(val.Elem().Type())
    h.Elem().FieldByName("Handler").Set(reflect.ValueOf(handlerVal))
    handler := h.Interface().(Handler)
    err = handler.Before()
    if err != nil {
      StatusPage(handler, http.StatusInternalServerError, err.Error())
      return
    }
    switch strings.ToUpper(request.Method) {
      case "POST":
        err = handler.Post()
      case "GET", "HEAD":
        err = handler.Get()
    }
    if err != nil {
      err = StatusPage(handler, http.StatusInternalServerError, err.Error())
      return
    } else if !handlerVal.isStarted {
      handler.Start()
    }
    err = handler.After()
    if err != nil {
      StatusPage(handler, http.StatusInternalServerError, err.Error())
      return
    }
    return
  }
  NotFound(handlerVal)
  return
}
