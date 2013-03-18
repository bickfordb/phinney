package phinney

import (
  "go/ast"
  "go/parser"
  "go/token"
  "os"
  "os/exec"
  "path/filepath"
  "regexp"
  "strings"
  "text/template"
)

type Project struct {
  Root string
  Env map[string]string
}

func NewProject(root string) (p *Project, err error) {
  project := &Project{}
  root, err = filepath.Abs(root)
  if err != nil { return }
  project.Root = root
  project.Env = make(map[string]string)
  for _, item := range os.Environ() {
    parts := strings.SplitN(item, "=", 2)
    if len(parts) == 2 {
      key := parts[0]
      val := parts[1]
      project.Env[key] = val
    }
  }
  err = safemkdir(project.BuildDir())
  if err != nil { return }
  err = safemkdir(project.BuildSrcDir())
  if err != nil { return }
  project.Env["GOPATH"] = project.BuildDir() + ":" + project.Root + ":" + project.Env["GOPATH"]
  p = project
  return
}

func (p *Project) BuildSrcDir() string {
  return filepath.Join(p.BuildDir(), "src")
}

type Symbol struct {
  Package string
  File string
  Name string
}

type Visitor func (symbol *Symbol)

func dirs(path string) (result []string) {
  result = make([]string, 0)
  f, err := os.Open(path)
  if err != nil { return }
  result, _ = f.Readdirnames(0)
  return
}

func Declarations(root string, visitor Visitor) (err error) {
  srcDir := filepath.Join(root, "src")
  for _, pkg := range dirs(srcDir) {
    pkgDir := filepath.Join(root, "src", pkg)
    fset := token.NewFileSet()
    log.Debug("parse dir", pkgDir)

    nameToPackage, parseDirErr := parser.ParseDir(fset, pkgDir, func(fi os.FileInfo) (f bool) {
      f = !fi.IsDir()
      return
    }, 0)
    if parseDirErr != nil { err = parseDirErr; return }
    for packageName, aPackage := range nameToPackage {
      for _, aFile := range aPackage.Files {
        for _, decl := range aFile.Decls {
          switch t := decl.(type) {
          case *ast.FuncDecl:
            visitor(&Symbol{packageName, aFile.Name.Name, t.Name.Name})
          }
        }
      }
    }
  }
  return
}

func Exists(path string) bool {
  _, err := os.Stat(path)
  return err == nil
}

func isDir(path string) bool {
  fileInfo, err := os.Stat(path)
  return err == nil && fileInfo.IsDir()
}

func ProjectDir(at string) string {
  at, err := filepath.Abs(at)
  if err != nil { return "" }
  for {
    if Exists(at + "/src") && isDir(at + "/src") {
        return at
    }
    at, _ = filepath.Split(at)
  }
  return ""
}

func (p *Project) FindPackagesContainingSymbol(pat *regexp.Regexp) (result []string, err error) {
  log.Debug("src dir: %s", p.SrcDir())
  err = Declarations(p.Root, func(sym *Symbol) {
    if pat.MatchString(sym.Name) {
      result = append(result, sym.Package)
    }
    return
  })
  if err != nil { return }
  return
}

func pathSegments(s string) (ret []string) {
  ret = strings.Split(s, "/")
  if len(ret) > 0 && ret[0] == "" {
    ret = ret[1:]
  }
  return
}

type Resource struct {
  Path string
  SrcPath string
}

func (project *Project) SrcDir() (result string) {
  return filepath.Join(project.Root, "src")
}

func (project *Project) BuildDir() (result string) {
  return filepath.Join(project.Root, ".build")
}

func (project *Project) Resources() (result []Resource, err error) {
  srcDir := project.SrcDir()
  srcParts := pathSegments(srcDir)
  filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) (err error) {
    if info.IsDir() { return }
    ext := filepath.Ext(path)
    if ext == ".go" { return }
    pkgParts := pathSegments(path)
    pkgParts = pkgParts[len(srcParts):]
    if len(pkgParts) < 3 { return }
    if pkgParts[1] != "templates" && pkgParts[1] != "assets" { return }
    dstPath := filepath.Join(pkgParts...)
    result = append(result, Resource{dstPath, path})
    return
  })
  return
}

func (b *Project) WriteTo(path, source string, templateData interface {}) (err error) {
  path = filepath.Join(b.BuildSrcDir(), path)
  log.Debug("write to: %s", path)
  parent := filepath.Dir(path)
  err = os.MkdirAll(parent, 0744)
  switch {
  case err != nil && os.IsExist(err):
    err = nil
  case err != nil:
    return
  }
  tmpl, err := template.New(path).Parse(source)
  if err != nil { return }
  dst, err := os.Create(path)
  if err != nil { return }
  err = tmpl.Execute(dst, templateData)
  if err != nil { return }
  dst.Close()
  return
}

func (p *Project) Exec(args ...string) (process *os.Process, err error) {
  executable := args[0]
  executable, err = exec.LookPath(executable)
  if err != nil { return }
  attrs := &os.ProcAttr{}
  attrs.Env = make([]string, 0)
  for key, val := range p.Env {
    attrs.Env = append(attrs.Env, key + "=" + val)
  }
  attrs.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
  process, err = os.StartProcess(executable, args, attrs)
  return
}

