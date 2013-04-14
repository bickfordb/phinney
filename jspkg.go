package phinney

import (
  _ "os"
  _ "os/exec"
  "bytes"
  "path/filepath"
  "fmt"
  "regexp"
  "errors"
)

var requirePat *regexp.Regexp = regexp.MustCompile(`require[(]\s*["](.*)["][)]`)

func (b *bundler) processRequire(name string) (err error) {
  if b.visited[name] { return }
  b.visited[name] = true
  var srcBytes []byte
  for _, includePath := range b.includePaths {
    path := filepath.Join(includePath, name + ".js")
    println("check path", path)
    srcBytes = Resources.Get(path)
    if srcBytes != nil { break }
  }
  if srcBytes == nil {
    err = fmt.Errorf("cant find module %q", name)
    return
  }
  if (b.dst.Len() == 0) {
    b.dst.WriteString(`if (typeof window == "undefined") window = {};` + "\n");
    b.dst.WriteString(`if (typeof window._modules == "undefined") window._modules = {};` + "\n")
    b.dst.WriteString(`if (typeof window.require == "undefined") window.require = function(m) { return window._modules[m]; };` + "\n")
  }
  b.dst.WriteString("(function(){\n")
  b.dst.WriteString("var exports = {};\n")
  b.dst.WriteString("window._modules[\"" + name + "\"] = exports;\n")
  src := bytes.NewBuffer(srcBytes)
  for src.Len() != 0 {
    line, _ := src.ReadString('\n')
    println(line)
    if m := requirePat.FindStringSubmatch(line); m != nil {
      err = b.processRequire(m[1])
      if err != nil { return err }
    }
    b.dst.WriteString(line)
  }
  b.dst.WriteString("})();\n")
  return
}

type bundler struct {
  includePaths []string
  visited map[string]bool
  dst *bytes.Buffer
}

func processBundle(module string) (result []byte, err error) {
  b := newBundler()
  err = b.processRequire(module)
  if err != nil { return}
  result = b.dst.Bytes()
  return
}

func newBundler() *bundler {
  b := &bundler{}
  b.visited = make(map[string]bool)
  b.includePaths = make([]string, 0)
  b.dst = &bytes.Buffer{}
  return b
}

var notFoundError = errors.New("not found")

func (app *App) JSBundle(rootPattern string, includePaths []string) {
  app.Get(rootPattern, func(req *Request) (err error) {
    if len(req.Args) != 1 {
      req.NotFound()
      return
    }
    moduleName := req.Args[0]
    b := newBundler()
    b.includePaths = includePaths
    err = b.processRequire(moduleName)
    if err != nil && err == notFoundError {
      req.NotFound()
      err = nil
      return
    } else if err != nil {
      return
    }
    req.Response.Header().Add("Content-Type", "application/javascript")
    req.Response.Write(b.dst.Bytes())
    return
  })
}




