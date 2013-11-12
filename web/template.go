package web

import (
  "bytes"
  "io"
  "path/filepath"
  "html/template"
)

func SendTemplate(h Handler, templatePath string, data interface{}) (err error) {
  f, err := OpenResource(templatePath, nil)
  if err != nil {
    return
  }
  t := template.New("")
  t.Delims("{%", "%}")
  if h.App().templateFuncs != nil {
    t = t.Funcs(h.App().templateFuncs)
  }
  t, err = t.ParseFiles(f.Name())
  if err != nil {
    return
  }
  if t == nil {
    panic("expecting a template")
  }
  var out bytes.Buffer
  err = t.ExecuteTemplate(&out, filepath.Base(f.Name()), data)
  if err != nil {
    return
  }
  io.Copy(h, &out)
  return
}

