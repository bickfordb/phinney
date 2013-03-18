package phinney

import (
  "os"
  "fmt"
  "bytes"
)

func safemkdir(p string) (err error) {
  mkDirErr := os.MkdirAll(p, 0755)
  if mkDirErr != nil && !os.IsExist(mkDirErr) {
    err = mkDirErr
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
