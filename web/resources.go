package web

import (
  "fmt"
  "os"
  "strings"
  "path/filepath"
)

var defaultResourcePaths []string = []string{"."}

func init() {
  for _, tok := range strings.Split(os.Getenv("RESOURCE_PATH"), ":") {
    if tok != "" {
      defaultResourcePaths = append(defaultResourcePaths, tok)
    }
  }
}

func OpenResource(path string, inPaths []string) (f *os.File, err error) {
  if len(path) > 0 && path[0] == '/' {
    f, err = os.Open(path)
    return
  }
  if inPaths == nil {
    inPaths = defaultResourcePaths
  }
  for _, p := range inPaths {
    f, err = os.Open(filepath.Join(p, path))
    if err != nil {
      if os.IsNotExist(err) {
        err = nil
      } else {
        return
      }
    }
    return
  }
  err = fmt.Errorf("cant find %q", path)
  return
}

