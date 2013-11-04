package web

import (
  "bytes"
  "io"
  "html/template"
)

func SendTemplate(h Handler, templatePath string, data interface{}) (err error) {
  f, err := OpenResource(templatePath, nil)
  if err != nil {
    return
  }
  t, err := template.ParseFiles(f.Name())
  if err != nil {
    return
  }
  var out bytes.Buffer
  err = t.Execute(&out, data)
  if err != nil {
    return
  }
  io.Copy(h, &out)
  return
}

