package phinney

import (
  "compress/zlib"
  "bytes"
  "io"
)

type ResourceData []byte

type resources struct {
  data map[string]ResourceData
}


var Resources = &resources{data: make(map[string]ResourceData)}

func (r *resources) Define(path string, data []byte) {
  r.data[path] = EncodeResourceData(data)
}

func (r *resources) Get(path string) (result []byte) {
  src, ok := r.data[path]
  if !ok { return }
  return DecodeResourceData(src)
}

func (r *resources) Keys() (result []string) {
  result = make([]string, 0, len(r.data))
  for k, _ := range r.data {
    result = append(result, k)
  }
  return
}

func (r *resources) DefineRaw(path string, data ResourceData) {
  r.data[path] = data
}

func EncodeResourceData(data []byte) ResourceData {
  dstBuf := bytes.NewBuffer(make([]byte, 0, len(data)))
  w := zlib.NewWriter(dstBuf)
  w.Write(data)
  w.Close()
  return dstBuf.Bytes()
}

func DecodeResourceData(src []byte) []byte {
  reader, err := zlib.NewReader(bytes.NewBuffer(src))
  if err != nil {
    log.Debug("err: ", err.Error())
    return nil
  }
  dstBuf := bytes.NewBuffer(make([]byte, 0, len(src)))
  io.Copy(dstBuf, reader)
  return dstBuf.Bytes()
}


