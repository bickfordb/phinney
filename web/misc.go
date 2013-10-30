package web

import (
  "net/http"
)

func (r *Handler) TemporaryRedirect(location string) (err error) {
	r.Header().Add("Location", location)
	r.StatusPage(http.StatusTemporaryRedirect, "")
	return
}

func (r *Handler) Unauthorized() (err error) {
	r.StatusPage(http.StatusUnauthorized, "")
	return
}

func (r *Handler) BadRequest(anError error) (err error) {
  msg := ""
  if err != nil {
    msg = anError.Error()
  }
  r.StatusPage(http.StatusBadRequest, msg)
	return
}
