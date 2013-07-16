package phinney

import (
  "encoding/json"
  "fmt"
  "regexp")

var JSONContentType *regexp.Regexp = regexp.MustCompile(`^.*application[/]json\s*(?:[;].+)$`)

func (r *Request) JSONParams(record interface {}) (e error) {
  contentType := r.Request.Header.Get("Content-Type")
  if JSONContentType.MatchString(contentType) {
    e = fmt.Errorf("expecting JSON Content-Type but got %q", contentType)
    return
  }
  dec := json.NewDecoder(r.Request.Body)
  e = dec.Decode(record)
  return
}

func (r *Request) JSONResponse(record interface {}) (err error) {
  r.Response.Header().Set("Content-Type", "application/json")
  enc := json.NewEncoder(r.Response)
  err = enc.Encode(record)
  return
}


func (r *Request) TemporaryRedirect(location string) (err error) {
  r.Response.Header().Add("Location", location)
  r.Response.WriteHeader(307)
  return
}

func (r *Request) Unauthorized() (err error) {
  r.Response.WriteHeader(401)
  return
}

