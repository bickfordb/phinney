package phinney

import (
  "bytes"
  "fmt"
  "html/template"
  "io"
  "regexp"
  "strings"
)

/*
Due to the nature of Go templates, dependencies must be parsed before their parent is parsed.  A common pattern is to have a layout template

$pkg/templates/$x
$pkg/templates/helpers/$y
*/

type templates struct {
  helperTemplates *template.Template
  templates map[string]*template.Template
}

var Templates templates

func init() {
  Templates.helperTemplates = template.Must(template.New("---").Parse(""))
  Templates.templates = make(map[string]*template.Template)
}

func (tmpls *templates) RegisterTemplate(path string, source string) (err error) {
  t, err := tmpls.helperTemplates.Clone()
  if err != nil { return }
  t = t.New(path)
  t, err = t.Parse(source)
  if err != nil { return }
  tmpls.templates[path] = t
  return
}

func (tmpls *templates) RegisterHelperTemplate(path string, source string) (err error) {
  t, err :=  tmpls.helperTemplates.New(path).Parse(source)
  if err != nil {
    return
  }
  tmpls.helperTemplates = t
  return
}

func (tmpls *templates) Template(writer io.Writer, name string, data interface{}) (err error) {
  t, exists := tmpls.templates[name]
  if !exists {
    err = fmt.Errorf("missing template %s", name)
    return
  }
  err = t.Execute(writer, data)
  return
}

func (req *Request) Template(name string, data interface{}) (err error) {
  buf := bytes.NewBuffer(make([]byte, 0, 10000))
  err = Templates.Template(buf, name, data)
  if err != nil { return }
  req.Response.Header().Add("Content-Type", "text/html; charset=utf-8")
  req.Response.WriteHeader(200)
  w := bytes.NewBuffer(buf.Bytes())
  _, err = w.WriteTo(req.Response)
  return
}

func (tmpls *templates) LoadResourceTemplates() (err error) {
  templatesPat := regexp.MustCompile("^.+/templates/(.+)$")
  helpersPat := regexp.MustCompile("^.+/helpers/.+$")
  keys := Resources.Keys()
  for _, key := range keys {
    if !templatesPat.MatchString(key) || !helpersPat.MatchString(key) { continue }
    data := Resources.Get(key)
    if data == nil { continue }
    path := strings.Replace(key, "/templates/", "", -1)
    err = tmpls.RegisterHelperTemplate(path, string(data))
    if err != nil { return }
  }
  for _, key := range keys {
    if !templatesPat.MatchString(key) || helpersPat.MatchString(key) { continue }
    data := Resources.Get(key)
    if data == nil { continue }
    path := strings.Replace(key, "/templates/", "", -1)
    err = tmpls.RegisterTemplate(path, string(data))
    if err != nil { return }
  }
  return
}


