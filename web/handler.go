package web

import (
  "net/http"
)

type Handler interface {
  BeforeRequest() error
  ResponseWriter() http.ResponseWriter
  AfterRequest() error
  Head() error
  Post() error
  Delete() error
  Get() error
  Trace() error
  Request() *http.Request
  Args() map[string]string
  Status() int
  SetStatus(code int)
  SetIsChunked(chunked bool)
  IsChunked() bool
  Header() http.Header
  Start()
  NoteStarted()
  Write(buf []byte) (n int, err error)
  App() *App
}

type concreteHandler struct {
  request *http.Request
  args map[string]string
  status int
  isFinished bool
  isStarted bool
  isChunked bool
  response http.ResponseWriter
  app *App
}

func (c *concreteHandler) App() *App {
  return c.app
}

func (c *concreteHandler) IsChunked() bool {
  return c.isChunked
}

func (c *concreteHandler) SetIsChunked(t bool) {
  c.isChunked = t
}

func (c *concreteHandler) BeforeRequest() (err error) {
  return
}
func (c *concreteHandler) AfterRequest() (err error) {
  return
}

func (c *concreteHandler) ResponseWriter() http.ResponseWriter {
  return c.response
}

func (c *concreteHandler) Get() (err error) {
  StatusPage(c, http.StatusMethodNotAllowed, "")
  return
}

func (c *concreteHandler) Post() (err error) {
  StatusPage(c, http.StatusMethodNotAllowed, "")
  return
}

func (c *concreteHandler) Head() (err error) {
  StatusPage(c, http.StatusMethodNotAllowed, "")
  return
}

func (c *concreteHandler) Trace() (err error) {
  StatusPage(c, http.StatusMethodNotAllowed, "")
  return
}

func (c *concreteHandler) Delete() (err error) {
  StatusPage(c, http.StatusMethodNotAllowed, "")
  return
}

func (c *concreteHandler) Put() (err error) {
  StatusPage(c, http.StatusMethodNotAllowed, "")
  return
}

func (c *concreteHandler) Request() *http.Request {
  return c.request
}

func (c *concreteHandler) Args() map[string]string {
  return c.args
}

func (c *concreteHandler) Status() int {
  return c.status
}

func (c *concreteHandler) SetStatus(code int) {
  c.status = code
}


func (h *concreteHandler) Write(buf []byte) (n int, err error) {
  h.Start()
  n, err = h.response.Write(buf)
  return
}


func (h *concreteHandler) Header() http.Header {
  return h.response.Header()
}

func (h *concreteHandler) NoteStarted() {
  h.isStarted = true
}

func (h *concreteHandler) Start() {
  if h.isStarted {
    return
  }
  h.response.WriteHeader(h.status)
  h.isStarted = true
}

