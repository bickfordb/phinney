package phinney

import (
	"bytes"
	"fmt"
	"mime"
	"io"
	"os"
	"path/filepath"
  "regexp"
	"time"
  "archive/zip"
)
const modifiedFormat = "Thu, 02 Jan 2006 15:04:05 MST"

type Resources struct {
	Stores []Store
}

func NewResources(resourcePaths []string) (result *Resources) {
	result = &Resources{}
  result.Stores = append(result.Stores, memory)
	for _, p := range resourcePaths {
		result.Stores = append(result.Stores, &FS{p})
	}
	return
}

type Store interface {
	Keys() []string
	Get(key string) []byte
	LastModified(key string) *time.Time
}

type MemoryEntry struct {
  Data []byte
  ModTime time.Time
}

type Memory struct {
  Entries map[string]MemoryEntry
}

func NewMemory() *Memory {
  m := &Memory{}
  m.Entries = make(map[string]MemoryEntry)
  return m
}

func (m *Memory) Keys() (result []string) {
  for k, _ := range m.Entries {
    result = append(result, k)
  }
  return
}

func (m *Memory) Get(key string) []byte {
  v, exists := m.Entries[key]
  if !exists {
    return nil
  } else {
    return v.Data
  }
}

func (m *Memory) LastModified(key string) *time.Time {
  v, exists := m.Entries[key]
  if !exists {
    return nil
  } else {
    return &v.ModTime
  }
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
  defer aFile.Close()
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

func (f *FS) LastModified(key string) (result *time.Time) {
	p := filepath.Join(f.Root, key)
	aFile, err := os.Open(p)
	if err != nil {
		return
  }
  defer aFile.Close()
  info, err := aFile.Stat()
  if err != nil {
    return
  }
  t := info.ModTime()
  result = &t
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

func (r *Resources) LastModified(path string) (result *time.Time) {
  path = filepath.Clean(path)
	for _, store := range r.Stores {
		result = store.LastModified(path)
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

func init() {
  memory = NewMemory()
  r, err := zip.OpenReader(os.Args[0])
  if err != nil {
    return
  }
  defer r.Close()
  for _, f := range r.File {
    h, e := f.Open()
    if e != nil {
      continue
    }
    b := &bytes.Buffer{}
    b.Grow(int(f.UncompressedSize64))
    io.Copy(b, h)
    entry := MemoryEntry{}
    entry.Data = b.Bytes()
    entry.ModTime = f.ModTime()
    memory.Entries[f.Name] = entry
    h.Close()
  }
}

var memory *Memory = nil

func (app *App) Resource(pattern string, resourcePath string) {
	app.Router.Add(&Route{
		Pattern: regexp.MustCompile(pattern),
		Handler: func(req *Request) (err error) {
			if len(req.Args) != 1 {
				req.NotFound()
				return
			}
			path := fmt.Sprintf(resourcePath, req.Args[0])
      mtime := app.Resources.LastModified(path)
      if mtime == nil {
        req.NotFound()
        return
      }
      utcMTime := (*mtime).UTC()

      ifModifiedSinceS := req.Request.Header.Get("If-Modified-Since")
      if ifModifiedSinceS != "" {
        var ifModifiedSince time.Time
        ifModifiedSince, err = time.Parse(modifiedFormat, ifModifiedSinceS)
        if err != nil {
          return
        }
        ifModifiedSince = ifModifiedSince.UTC()
        if (ifModifiedSince.Equal(utcMTime) || ifModifiedSince.After(utcMTime)) {
          req.Response.WriteHeader(304)
          return
        }
      }

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
        req.Response.Header().Add("Last-Modified", utcMTime.Format(modifiedFormat))
				req.Response.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
				req.Response.Write(data)
			}
			return
		}})
}
