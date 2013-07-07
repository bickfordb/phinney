package phinney

import (
	"net/http"
  "fmt"
  "io"
)

func (req *Request) NotFound() {
	http.NotFound(req.Response, req.Request)
}

func writeAll(data []byte, w io.Writer) (err error) {
  var amt int
  var e error
  i := 0
  for i < len(data) && e == nil {
    amt, e = w.Write(data[i:])
    i += amt
  }
  if e != io.EOF && e != nil {
    err = e
  }
  return
}

func ServeBytes (data []byte, headers map[string]string) Handler {
  return func(req *Request) (err error) {
    // FIXME, handle HEAD, Conditional GET
    for key, val := range headers {
      req.Response.Header().Add(key, val)
    }
    req.Response.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
    err = writeAll(data, req.Response)
    return
  }
}
