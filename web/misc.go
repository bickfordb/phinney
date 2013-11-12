package web

import (
  "bytes"
  "net/http"
  "fmt"
  "html"
  "io"
  "io/ioutil"
)

func WriteHTML(h Handler, s string) (err error) {
  h.Header().Set("Content-Type", "text/html")
  _, err = io.Copy(h, bytes.NewBuffer([]byte(s)))
  return
}

func TemporaryRedirect(r Handler, location string) (err error) {
	r.Header().Add("Location", location)
	StatusPage(r, http.StatusTemporaryRedirect, "")
	return
}

func Unauthorized(r Handler) (err error) {
	StatusPage(r, http.StatusUnauthorized, "")
	return
}

func BadRequest(r Handler, anError error) (err error) {
  msg := ""
  if err != nil {
    msg = anError.Error()
  }
  StatusPage(r, http.StatusBadRequest, msg)
	return
}

func NotFound(h Handler) (err error) {
  return StatusPage(h, http.StatusNotFound, "")
}


func StatusPage(h Handler, code int, msg string) (err error) {
  h.SetStatus(code)
  body := fmt.Sprintf(
    `<html><body><h1>%d %s</h1><pre>%s</pre></body></html>`,
    code, http.StatusText(code), html.EscapeString(msg))
  err = WriteHTML(h, body)
  return
}

type DummyResponseWriter struct {
  Headers http.Header
  *bytes.Buffer
  Status int
}

func NewDummyResponseWriter() *DummyResponseWriter {
  return &DummyResponseWriter{
    Headers: make(http.Header),
    Buffer: &bytes.Buffer{},
  }
}

func (d *DummyResponseWriter) Header() http.Header {
  return d.Headers
}

func (d *DummyResponseWriter) WriteHeader(status int) {
  d.Status = status
}

func (d *DummyResponseWriter) HTTPResponse(request *http.Request) *http.Response {
  return &http.Response{
    Status: http.StatusText(d.Status),
    StatusCode: d.Status,
    Header: d.Headers,
    Request: request,
    Body: ioutil.NopCloser(d.Buffer),
  }
}

func ClearCookie (h Handler, name string) {
  h.Header().Add("Set-Cookie", name + "=deleted; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT")
}
