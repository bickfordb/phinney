package web

import (
  "fmt"
  "bytes"
  "io"
  "net/http"
  "html"
)

type httpHandler interface {
  IsHTTPHandler()
}

type IsHead interface {
  Head() error
}

type IsGet interface {
  Get() error
}

type IsPost interface {
  Post() error
}

type IsDelete interface {
  Delete() error
}

type IsPut interface {
  Put() error
}

type IsTrace interface {
  Trace() error
}

type IsBefore interface {
  Before() error
}

type IsAfter interface {
  After() error
}

type Handler struct {
  Request *http.Request
  Args map[string]string
  Status int
  isFinished bool
  isStarted bool
  IsChunked bool
  response http.ResponseWriter
  app *App
}

func (h *Handler) WriteHTML(s string) (err error) {
  h.Header().Set("Content-Type", "text/html")
  _, err = io.Copy(h, bytes.NewBuffer([]byte(s)))
  return
}

func (h *Handler) Write(buf []byte) (n int, err error) {
  h.Start()
  n, err = h.response.Write(buf)
  return
}

func (h *Handler) StatusPage(code int, msg string) (err error) {
  h.Status = code
  body := fmt.Sprintf(
    `<html><body><h1>%d %s</h1><pre>%s</pre></body></html>`,
    code, http.StatusText(code), html.EscapeString(msg))
  err = h.WriteHTML(body)
  return
}

func (h *Handler) Header() http.Header {
  return h.response.Header()
}

func (h *Handler) IsHTTPHandler() {
}

func (h *Handler) Start() {
  if h.isStarted {
    return
  }
  h.response.WriteHeader(h.Status)
  h.isStarted = true
}

func (h *Handler) NotFound() (err error) {
  return h.StatusPage(http.StatusNotFound, "")
}

