package phinney

import "fmt"

type subErr struct {
  pkg string
  err error
}

func (e *subErr) Error() string {
  return fmt.Sprintf("%s: %s", e.pkg, e.err.Error())
}

func subError(pkg string, err error) error {
  return &subErr{pkg, err}
}

