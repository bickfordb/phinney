package phinney

import (
  "flag"
  "path/filepath"
  "fmt"
  "regexp"
)

type ProjectConf struct {
  Title string
  Apps []string
}

func newProject() (err error) {
  projectName := flag.Arg(1)
  pat := regexp.MustCompile("^([a-z]+[A-Z0-9_]+)$")
  if !pat.MatchString(projectName) {
    fmt.Errorf("expecting a project name")
  }
  projectDir, _ := filepath.Abs(projectName)
  log.Debug("project dir: %s", projectDir)
  safemkdir(projectDir)

  srcDir := filepath.Join(projectDir, "src")
  safemkdir(srcDir)

  conf := &ProjectConf{}
  conf.Apps = append(conf.Apps, projectName)
  conf.Title = projectName
  log.Debug("writing conf: %+v", conf)
  confPath := filepath.Join(projectDir, "project.json")
  writeJSON(confPath, conf)
  routePath := filepath.Join(projectDir, "src", projectName, "route.go")

  var routeContext struct {
    PackageName string
  }
  routeContext.PackageName = projectName
  err = writeTemplate(routePath, routeTemplate, routeContext)
  if err != nil { return }
  return
}

var routeTemplate string = `
package {{.PackageName}}

import (
  "phinney"
)

func SetupRoutes(router *phinney.Router) {
  router.GET("^/somewhere",
}
`

