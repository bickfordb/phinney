package web

import (
  "regexp"
)


func (app *App) Route(path string, handler httpHandler) {
  var route route
  route.pathTemplate = parseURLPattern(path)
  route.handler = handler
  app.routes = append(app.routes, route)
}

type route struct {
  pathTemplate pathTemplate
  handler httpHandler
}

type pathTemplate struct {
  original string
  pattern *regexp.Regexp
  variables []string
}

func (r *route) matches(path string) (matches bool, args map[string]string) {
  submatches := r.pathTemplate.pattern.FindAllStringSubmatch(path, -1)
  if submatches == nil {
    return
  }
  matches = true
  args = make(map[string]string)
  for idx, variable := range r.pathTemplate.variables {
    args[variable] = submatches[0][idx]
  }
  return
}

var varPattern = regexp.MustCompile("/[:]?[^/]*")
/*
parse path like "/photos/:id/something" to =>
  (pattern, [strings])
*/

func parseURLPattern(url string) (result pathTemplate) {
  result.original = url
  result.variables = make([]string, 0)
  pattern := "^"
  for _, submatch := range varPattern.FindAllStringSubmatch(url, -1) {
    s := submatch[0][1:]
    pattern += "/"
    if len(s) > 1 && s[1] == ':' {
      result.variables = append(result.variables, s[2:])
      pattern += "([^/]+)"
    } else if len(s) > 0 {
      pattern += regexp.QuoteMeta(s[1:])
    }
  }
  pattern += "$"
  println("pattern", pattern)
  result.pattern = regexp.MustCompile(pattern)
  return
}

