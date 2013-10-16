package phinney

import "fmt"
import "runtime"

type NestedError struct {
  File string
  Line int
  Err error
}

func (e *NestedError) Error() string {
  msg := ""
  if e.Err != nil {
    msg = e.Err.Error()
  }
  return fmt.Sprintf("%s (%d): %s", e.File, e.Line, msg)
}

func NestError(err error) (e error) {
  if err == nil {
    return
  }
  if err != nil {
    e = err
  }
  _, file, line, _ := runtime.Caller(1)
  file = ""
  line = 0
  e = &NestedError{file, line, err}
  return
}

