package phinney

import (
  "encoding/json"
  "regexp")

var JSONContentType *regexp.Regexp = regexp.MustCompile(`^application[/]json.*$`)

func (r *Request) IsJSON() bool {
  return JSONContentType.MatchString(r.Request.Header.Get("Content-Type"))

}
func (r *Request) JSONParams(record interface {}) (e error) {
  if r.Request.Method == "GET" {
    data := r.Request.URL.Query().Get("params")
    if (data != "") {
      log.Debug("url: %q", r.Request.URL.String())
      log.Debug("decoding %q", data)
      json.Unmarshal([]byte(data), record)
      // ignore GET-ish errors
    }
  } else {
    if !r.IsJSON() {
      //e = fmt.Errorf("expecting JSON Content-Type but got %q", r.Request.Header.Get("Content-Type"))
      return
    }
    dec := json.NewDecoder(r.Request.Body)
    e = dec.Decode(record)
  }
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

func (r *Request) BadRequest() (err error) {
  r.Response.WriteHeader(400)
  return
}
