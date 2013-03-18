package phinney

import "net/http"

type URLPattern struct {
  Pattern string
  Handler http.Handler
}
