package phinney

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func (app *App) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
  err := app.serveHTTP(writer, request)
  if err != nil {
    app.RenderError(writer, request, err)
  }
  return
}


func (app *App) serveHTTP(writer http.ResponseWriter, request *http.Request) (err error) {

  req := &Request{
			Request:  request,
			Response: writer,
			App:      app,
			Context:  make(map[string]interface{})}
  route, args := app.Router.Search(request.Method, request.URL.Path)
  req.Args = args
  var halt bool
  for _, plugin := range app.Plugins {
    if halt {
      break
    }
    if plugin.StartRequest != nil {
      halt, err = plugin.StartRequest(app, req)
      if err != nil {
        return
      }
    }
  }
  for _, plugin := range app.Plugins {
    if halt {
      break
    }
    if plugin.BeforeHandler != nil {
      halt, err = plugin.BeforeHandler(app, req)
      if err != nil {
        return
      }
    }
  }
  if !halt {
    if route != nil {
      err = route.Handler(req)
    } else {
      http.NotFound(writer, request)
    }
  }
  for _, plugin := range app.Plugins {
    if plugin.AfterResponse != nil {
      err = plugin.AfterResponse(app, req)
    }
  }
  return
}

func (app *App) RenderError(writer http.ResponseWriter, request *http.Request, err error) {
  log.Debug("Request Error: %q", err.Error())
  writer.WriteHeader(503)
  writer.Header().Add("Content-Type", "text/html; charset=utf-8")
  buf := &bytes.Buffer{}
  ErrorTemplate.Execute(buf, map[string]interface{}{"Error": err.Error()})
  writer.Header().Add("Content-Length", fmt.Sprintf("%d", len(buf.Bytes())))
  writer.Write(buf.Bytes())
}

var ErrorTemplate *template.Template = template.Must(template.New("error").Parse(
	`<!doctype html>
<html>
  <body>
    <h1>503 Server Error</h1>
    <p>
      <pre>{{.Error}}</pre>
    </p>
  </body>
</html>
`))

func (app *App) Listen(addr string) {
	go func() {
		log.Info("listening on %q", addr)
		server := &http.Server{
			Addr:           addr,
			Handler:        app,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20}
		e := server.ListenAndServe()
		app.Errors <- MakeAppError("Listen", e.Error())
	}()
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

type Request struct {
	Args     []string
	Request  *http.Request
	Response http.ResponseWriter
	App      *App
	Context  map[string]interface{}
}

