package phinney

import (
  "bytes"
  "compress/zlib"
  "io"
  "os"
  "regexp"
  "strings"
)

func (project *Project) buildWeb() (err error) {
  var context struct {
    Packages []string
  }
  binPath := project.BuildDir() + "/web"
  err = saferm(binPath)
  if err != nil { return }

  err = project.prepareResources()
  if err != nil { return }

  context.Packages, err = project.FindPackagesContainingSymbol(regexp.MustCompile("^SetupApp"))
  if err != nil { return }

  err = project.WriteTo("main.go", mainTemplate, context)
  if err != nil { return }

  proc, err := project.Exec("go", "build", "-o", binPath, project.BuildSrcDir() + "/main.go")
  proc.Wait()
  return
}

var mainTemplate = `
package main

import (
  _ "phinney_resources"
  "github.com/bickfordb/phinney"
  "github.com/bickfordb/lg"
  {{range $idx, $package := .Packages}}
  "{{$package}}"
  {{end}}
)

var log = lg.GetLog("web")

func main() {
  var err error = nil
  log.Info("starting")
  err = phinney.Templates.LoadResourceTemplates()
  log.Info("loaded templates")
  if err != nil { panic(err.Error()) }

  app := phinney.NewApp()
  {{range $idx, $package := .Packages}}
  {{$package}}.SetupApp(app)
  {{end}}
  log.Info("listening")
  err = app.Serve(":9090")
  if err != nil {
    log.Error("unexpected: %s", err.Error())
  }
}
`

var resourcesTemplate = `
package phinney_resources

import "github.com/bickfordb/phinney"

func init() {
  phinney.Resources.DefineRaw("{{.Path}}", {{.Literal}})
}

`
func compressed(path string) (result []byte, err error) {
  buf := bytes.NewBuffer(make([]byte, 0, 10000))
  zlibW := zlib.NewWriter(buf)
  srcFile, err := os.Open(path)
  if err != nil { return }
  _, err = io.Copy(zlibW, srcFile)
  if err != nil { return }
  err = zlibW.Close()
  if err != nil { return }
  result = buf.Bytes()
  return
}


func (p *Project) prepareResources() (err error) {
  resources, err := p.Resources()
  if err != nil { return }
  os.MkdirAll(p.BuildSrcDir() + "/phinney_resources", 0755)
  p.WriteTo("phinney_resources/blank.go", "package phinney_resources\n", nil)

  for _, resource := range resources {
    rsrcPath := "phinney_resources/" + strings.Replace(resource.Path, "/", "_", -1)  + ".go"
    var data struct {
      Literal string
      Path string
    }
    bs, rErr := compressed(resource.SrcPath)
    if rErr != nil { err = rErr; return }
    data.Literal = toByteLiteral(bs)
    data.Path = resource.Path
    err = p.WriteTo(rsrcPath, resourcesTemplate, data)
    if err != nil { return }
  }
  return
}

