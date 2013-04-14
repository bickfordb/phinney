package phinney

import (
  "os"
  "fmt"
  "bytes"
  "path/filepath"
  "encoding/json"
  "text/template"
  "strings"
)

func safemkdir(p string) (err error) {
  mkDirErr := os.MkdirAll(p, 0755)
  if mkDirErr != nil && !os.IsExist(mkDirErr) {
    err = mkDirErr
  }
  return
}

func safeMakeParentDirs(path string) (err error) {
  d := filepath.Dir(path)
  if d != "." {
    err = safemkdir(d)
  }
  return
}

func saferm(p string) (err error) {
  rmErr := os.Remove(p)
  if rmErr != nil && !os.IsNotExist(rmErr) {
    err = rmErr
  }
  return
}

func toByteLiteral(bs []byte) string {
  szEst := 10 + (5 * len(bs))
  buf := bytes.NewBuffer(make([]byte, 0, szEst))
  buf.WriteString("[]byte{")
  for i, b := range bs {
    if i != 0 {
      buf.WriteString(",")
    }
    buf.WriteString(fmt.Sprintf("0x%X", b))
  }
  buf.WriteString("}")
  return string(buf.Bytes())
}

func writeJSON(path string, data interface {}) (err error) {
  b, err := json.MarshalIndent(data, "", "  ")
  if err != nil { return }
  err = safeMakeParentDirs(path)
  if err != nil { return }
  aFile, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE, 0644)
  if err != nil { return }
  _, err = aFile.Write(b)
  if err != nil { return }
  _, err = aFile.Write([]byte("\n"))
  if err != nil { return }
  aFile.Close()
  return
}

func writeTemplate(path string, source string, data interface{}) (err error) {
  log.Debug("writing template %q", path)
  tmpl := template.Must(template.New(path).Parse(source))
  safeMakeParentDirs(path)
  f, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE, 0644)
  if err != nil { return }
  err = tmpl.Execute(f, data)
  return
}

func envDict() (result map[string]string) {
  result = make(map[string]string)
  for _, item := range os.Environ() {
    parts := strings.SplitN(item, "=", 2)
    if len(parts) == 2 {
      key := parts[0]
      val := parts[1]
      result[key] = val
    }
  }
  return
}
