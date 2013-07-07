package phinney

import (
	"bytes"
	"fmt"
	"mime"
	"os"
	"path/filepath"
  "regexp"
	"time"
)

type Resources struct {
	Stores []Store
	//data map[string]ResourceData
}

func NewResources(resourcePaths []string) (result *Resources) {
	result = &Resources{}
	for _, p := range resourcePaths {
		var store Store
		store = &FS{p}
		result.Stores = append(result.Stores, store)
	}
	return

}

type Store interface {
	Keys() []string
	Get(key string) []byte
	Timestamp(key string) time.Time
}

type FS struct {
	Root string
}

func (f *FS) Keys() (result []string) {
	return
}

func (f *FS) Get(key string) (result []byte) {
	p := filepath.Join(f.Root, key)
	aFile, err := os.Open(p)
	if err != nil {
		return
	}
	info, err := aFile.Stat()
	if err != nil {
		return
	}
	size := info.Size()

	data := make([]byte, 0, size)
	buf := bytes.NewBuffer(data)
	amtRead, err := buf.ReadFrom(aFile)
	if amtRead != size {
		return
	}
	if err != nil {
		return
	}
	result = buf.Bytes()
	return
}

func (f *FS) Timestamp(key string) (result time.Time) {
	return
}

func (r *Resources) Get(path string) (result []byte) {
	path = filepath.Clean(path)
	for _, store := range r.Stores {
		result = store.Get(path)
		if result != nil {
			return
		}
	}
	return
}

func (r *Resources) Keys() (result []string) {
	result = make([]string, 0)
	for _, s := range r.Stores {
		for _, key := range s.Keys() {
			result = append(result, key)
		}
	}
	return
}

//func (r *Resources) loadZipData(src string) (err error) {
//  err = nil
//  sz := exeLen(src)
//  if sz == 0 { return }
//  f, err := os.Open(src)
//  if err != nil { return err }
//  stat, err := f.Stat()
//  if err != nil { return err }
//  rdr := io.NewSectionReader(f, int64(sz), stat.Size() - int64(sz))
//  zipfile, err := zip.NewReader(rdr, rdr.Size())
//  if err != nil { return }
//  for _, f := range zipfile.File {
//    buf := make([]byte, f.UncompressedSize)
//    if f.UncompressedSize != 0 {
//      h, err := f.Open()
//      if err != nil { return err }
//      defer h.Close()
//      _, err = io.ReadFull(h, buf)
//      if err != nil { return err }
//    }
//    r.Define(f.Name, buf)
//  }
//  return
//}
//
//func LoadResources() (err error) {
//  exec := os.Args[0]
//  err = Resources.loadZipData(exec)
//  if err != nil { return }
//
//  env := envDict()
//  resourcesPath := env["RESOURCEPATH"]
//  resourcePaths := strings.SplitN(resourcesPath, ":", -1)
//  for _, resourcePath := range resourcePaths {
//    if resourcePath == "" { continue }
//    _, e := os.Open(resourcePath)
//    if e != nil {
//      resourcesLog.Error("cant open %q", e.Error())
//      continue
//    }
//    err = filepath.Walk(resourcePath, func(path string, info os.FileInfo, err error) (err0 error) {
//      if (info.IsDir()) { return }
//      f, e := os.Open(path)
//      if e != nil {
//        return
//      }
//      theBytes := make([]byte, info.Size())
//      _, e = f.Read(theBytes)
//      if e != nil { return }
//      path = path[len(resourcePath) + 1:]
//      resourcesLog.Debug("path: %q", path);
//      Resources.Define(path, theBytes)
//      return
//    })
//    if err != nil { return }
//  }
//  return
//}
//

func (app *App) Resource(pattern string, resourcePath string) {
	app.Router.Add(&Route{
		Pattern: regexp.MustCompile(pattern),
		Handler: func(req *Request) (err error) {
			if len(req.Args) != 1 {
				req.NotFound()
				return
			}
			path := fmt.Sprintf(resourcePath, req.Args[0])
			data := app.Resources.Get(path)
			if data == nil {
				req.NotFound()
				return
			} else {
				contentType := mime.TypeByExtension(filepath.Ext(path))
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				req.Response.Header().Add("Content-Type", contentType)
				req.Response.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
				req.Response.Write(data)
			}
			return
		}})
}
