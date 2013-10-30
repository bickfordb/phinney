package web

import (
  "fmt"
  "net/http"
  "reflect"
  "strings"
)

type App struct {
  routes []route
}

func (app *App) ServeHTTP(response http.ResponseWriter, request *http.Request) {
  var err error
  path := request.URL.Path
  handlerVal := &Handler{}
  handlerVal.Request = request
  handlerVal.response = response
  handlerVal.Status = http.StatusOK
  handlerVal.app = app
  for _, route := range app.routes {
    fmt.Println("route: ", route)
    matches, args := route.matches(path)
    if !matches {
      continue
    }
    handlerVal.Args = args
    println("found handler");
    val := reflect.ValueOf(route.handler)
    h := reflect.New(val.Elem().Type())
    h.Elem().FieldByName("Handler").Set(reflect.ValueOf(handlerVal))
    handler := h.Interface().(httpHandler)
    if m, ok := handler.(IsBefore); ok {
      err = m.Before()
    }
    if err != nil {
      handlerVal.StatusPage(http.StatusInternalServerError, err.Error())
      return
    }
    handled := false
    switch strings.ToUpper(request.Method) {
      case "POST":
        println("is post")
        switch m := handler.(type) {
        case IsPost:
          println("ran post")
          err = m.Post()
          handled = true
        }
      case "GET":
        if m, ok := handler.(IsGet); ok {
          err = m.Get()
          handled = true
        }
    }
    fmt.Println("handled", handled)
    fmt.Println("status", handlerVal.Status)
    if !handled {
      println("sending 405")
      err = handlerVal.StatusPage(http.StatusMethodNotAllowed, "")
      return
    } else if err != nil {
      err = handlerVal.StatusPage(http.StatusInternalServerError, err.Error())
      return
    } else if !handlerVal.isStarted {
      handlerVal.Start()
    }
    switch handler := handler.(type) {
    case IsAfter:
      err = handler.After()
      if err != nil {
        handlerVal.StatusPage(http.StatusInternalServerError, err.Error())
        return
      }
    }
    return
  }
  notFound(handlerVal)
  println("could not find handler!")
}

func notFound(h *Handler) {
  h.response.Header().Add("Content-Type", "text/html")
  h.response.WriteHeader(404)
  h.response.Write([]byte(h.app.notFoundBody()))

}

func (app *App) notFoundBody() (result string) {
  return `<html><body><h1>Not Found</h1></body></html>`
}


