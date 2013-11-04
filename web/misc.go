package web

import (
  "bytes"
  "net/http"
  "fmt"
  "html"
  "io"
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

