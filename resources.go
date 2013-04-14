package phinney

import (
  "bytes"
  "archive/zip"
  "compress/zlib"
  "path/filepath"
  "io"
  "os"
  "strings"
  "github.com/bickfordb/lg"
  "debug/macho"
)


func machoLen(path string) (result uint64) {
  f, err := os.Open(path)
  if err != nil { return }
  mo, err := macho.NewFile(f)
  if err != nil { return }
  defer mo.Close()
  for _, load := range mo.Loads {
    if seg, ok := load.(*macho.Segment); ok {
      var end uint64 = seg.Offset + seg.Filesz
      if end > result { result = end }
    }
  }
  return
}

func exeLen(s string) (result uint64) {
  result = machoLen(s)
  if result > 0 { return result }
  return
}

type ResourceData []byte

type resources struct {
  data map[string]ResourceData
}

var resourcesLog = lg.GetLog("phinney.resources")

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
    resourcesLog.Debug("err: ", err.Error())
    return nil
  }
  dstBuf := bytes.NewBuffer(make([]byte, 0, len(src)))
  io.Copy(dstBuf, reader)
  return dstBuf.Bytes()
}

func (r *resources) loadZipData(src string) (err error) {
  log.Error("hi")
  err = nil
  sz := exeLen(src)
  if sz == 0 { return }
  f, err := os.Open(src)
  if err != nil { return err }
  stat, err := f.Stat()
  if err != nil { return err }
  rdr := io.NewSectionReader(f, int64(sz), stat.Size() - int64(sz))
  zipfile, err := zip.NewReader(rdr, rdr.Size())
  if err != nil { return }
  for _, f := range zipfile.File {
    buf := make([]byte, f.UncompressedSize)
    if f.UncompressedSize != 0 {
      h, err := f.Open()
      if err != nil { return err }
      defer h.Close()
      _, err = io.ReadFull(h, buf)
      if err != nil { return err }
    }
    r.Define(f.Name, buf)
  }
  return
}

func LoadResources() (err error) {
  exec := os.Args[0]
  err = Resources.loadZipData(exec)
  if err != nil { return }

  env := envDict()
  resourcesPath := env["RESOURCEPATH"]
  resourcePaths := strings.SplitN(resourcesPath, ":", -1)
  for _, resourcePath := range resourcePaths {
    if resourcePath == "" { continue }
    _, e := os.Open(resourcePath)
    if e != nil {
      resourcesLog.Error("cant open %q", e.Error())
      continue
    }
    err = filepath.Walk(resourcePath, func(path string, info os.FileInfo, err error) (err0 error) {
      if (info.IsDir()) { return }
      f, e := os.Open(path)
      if e != nil {
        return
      }
      theBytes := make([]byte, info.Size())
      _, e = f.Read(theBytes)
      if e != nil { return }
      path = path[len(resourcePath) + 1:]
      resourcesLog.Debug("path: %q", path);
      Resources.Define(path, theBytes)
      return
    })
    if err != nil { return }
  }
  return
}

