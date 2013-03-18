package phinney

import (
  "fmt"
  "mime"
  "net/http"
  "path/filepath"
  "time"
  "regexp"
)

type Method string

const (
  Any Method = ""
  Post = "POST"
  Get = "GET"
  Delete = "DELETE"
  Put = "PUT"
  Head = "HEAD"
)

type Handler (func(req *Request) (err error))

type Route struct {
  method Method
  path *regexp.Regexp
  handler Handler
}

func (app *App) Add(method Method, pattern string, handler Handler) {
  app.routes = append(app.routes, &Route{method, regexp.MustCompile(pattern), handler})
}

func (app *App) Get(pattern string, handler Handler) {
  app.Add(Get, pattern, handler)
}

func (app *App) Post(pattern string, handler Handler) {
  app.Add(Post, pattern, handler)
}

func (req *Request) NotFound() {
  http.NotFound(req.Response, req.Request)
}

func (app *App) Resource(pattern string, resourcePath string) {
  app.Add(Get, pattern, func(req *Request) (err error) {
    if len(req.Args) != 1 {
      req.NotFound()
      return
    }
    path := fmt.Sprintf(resourcePath, req.Args[0])
    data := Resources.Get(path)
    if data == nil {
      req.NotFound()
      return
    } else {
      contentType := mime.TypeByExtension(filepath.Ext(path))
      if contentType == "" {
        contentType = "application/octet-stream"
      }
      req.Response.Header().Add("Content-Type", contentType)
      req.Response.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
      req.Response.Write(data)
    }
    return
  })
}

func (app *App) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
  path := request.URL.Path
  for _, route := range app.routes {
    matches := route.path.FindStringSubmatch(path)
    if matches != nil {
      req := &Request{
        Request:request,
        Response:writer,
        App:app,
        Args:matches[1:]}
      err := route.handler(req)
      if err != nil {
        log.Debug("error: %s", err.Error())
        writer.WriteHeader(503)
        writer.Header().Add("Content-Type", "text/html; charset=utf-8")
        writer.Write([]byte(`<html><body><h1>Server Error</h1></body></html>`))
      }
      return
    }
  }
  http.NotFound(writer, request)
}

func (app *App) Serve(addr string) (err error) {
  server := &http.Server{
    Addr: addr,
    Handler: app,
    ReadTimeout: 10 * time.Second,
    WriteTimeout: 10 * time.Second,
    MaxHeaderBytes: 1 << 20}
  err = server.ListenAndServe()
  return
}

type HandlerFn struct {
  F http.HandlerFunc
}

func (h *HandlerFn) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
  h.F(writer, request)
}

func FromFn(f http.HandlerFunc) http.Handler {
  return &HandlerFn{f}
}


func RunRoutes(address string, patterns []*URLPattern) {
}

type Request struct {
  Args []string
  Request *http.Request
  Response http.ResponseWriter
  App *App
}

