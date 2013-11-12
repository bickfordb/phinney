package web

import (
  "net/http"
  "path/filepath"
)

type AssetsHandler struct {
  Root string
  Handler
}

func (a *AssetsHandler) Get() (err error) {
  path := a.Args()["path"]
  path = filepath.Join(a.Root, path)
  if path == "" {
    return NotFound(a)
  }
  f, err := OpenResource(path, nil)
  if err != nil {
    return
  }
  info, err := f.Stat()
  if err != nil {
    return
  }
  http.ServeContent(a.ResponseWriter(), a.Request(), path, info.ModTime(), f)
  a.NoteStarted()
  return
}

