package phinney

import (
  _ "os"
  _ "os/exec"
  "bytes"
  "path/filepath"
  "html/template"
  "fmt"
)

var headerJS = `
if (typeof window == "undefined") {
  var window = {};
}
if (typeof bundle == "undefined") {
  window.bundle = {"modules": {}, "templates": {}};
}

if (typeof require == "undefined") {
  window.require = function(m) { return bundle.modules[m]; };
}
`

func (bundle *JSBundle) compileModules(dst *bytes.Buffer) (err error) {
  for _, module := range bundle.Modules {
    var srcBytes []byte
    for _, includePath := range bundle.ModuleIncludePaths {
      srcBytes = Resources.Get(filepath.Join(includePath, module + ".js"))
      if srcBytes != nil { break }
    }
    if srcBytes == nil {
      err = fmt.Errorf("cant find module %q", module)
      return
    }
    moduleJS := &bytes.Buffer{}
    moduleJS.Write(srcBytes)
    moduleJS.WriteString("//@ sourceURL=" + module + "\n");
    dst.WriteString("{\n")
    dst.WriteString("  var exports = {};\n")
    dst.WriteString("  bundle.modules[\"" + template.JSEscapeString(module) + "\"] = exports;\n")
    dst.WriteString(`  eval("`)
    dst.WriteString(template.JSEscapeString(moduleJS.String()))
    dst.WriteString(`");`)
    dst.WriteString("\n");
    dst.WriteString("}\n");
  }
  return
}

func (b *JSBundle) compileTemplates(out *bytes.Buffer) (err error) {
  for _, name := range b.Templates {
    var srcBytes []byte
    for _, path := range b.TemplateIncludePaths {
      srcBytes = Resources.Get(filepath.Join(path, name))
      if srcBytes != nil { break }
    }
    if srcBytes == nil {
      err = fmt.Errorf("cant load %s", name)
      return
    }
    src := string(srcBytes)
    out.WriteString(fmt.Sprintf("bundle.templates[\"%s\"] = \"%s\";\n",template.JSEscapeString(name), template.JSEscapeString(src)));
  }
  return
}

type JSBundle struct {
  ModuleIncludePaths []string
  TemplateIncludePaths []string
  Modules []string
  Templates []string
}

func (bundle *JSBundle) compileHeader(dst *bytes.Buffer) (err error) {
  dst.WriteString(headerJS)
  return
}


func (bundle *JSBundle) compile(dst *bytes.Buffer) (err error) {
  err = bundle.compileHeader(dst)
  if err != nil { return }
  err = bundle.compileModules(dst)
  if err != nil { return }
  err = bundle.compileTemplates(dst)
  return
}

func (b *JSBundle) Handler() (handler Handler, err error) {
  out := &bytes.Buffer{}
  err = b.compile(out)
  if err != nil { return }
  handler = func(req *Request) (err error) {
    req.Response.Header().Add("Content-Type", "application/javascript")
    req.Response.Header().Add("Content-Length", fmt.Sprintf("%d", out.Len()))
    req.Response.Write(out.Bytes())
    return
  }
  return
}




