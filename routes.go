package phinney

import (
	"container/list"
	"sync"
  "regexp"
)

type Method string

const (
	Any    Method = ""
	Post          = "POST"
	Get           = "GET"
	Delete        = "DELETE"
	Put           = "PUT"
	Head          = "HEAD"
)

type Handler func(req *Request) (err error)

type Router struct {
	routes      *list.List
	nameToRoute map[string]*Route
	lock        sync.Mutex
}

func NewRouter() *Router {
	router := &Router{}
	router.routes = list.New()
	router.nameToRoute = make(map[string]*Route)
	return router
}

type Route struct {
	Method  Method
	Pattern *regexp.Regexp
	Handler Handler
	Name    string
	Reverse string
}

func (router *Router) Search(method string, path string) (result *Route, pathArgs []string) {
  router.lock.Lock()
  defer router.lock.Unlock()
  for i := router.routes.Front(); i != nil; i = i.Next() {
    route := i.Value.(*Route)
    if route.Method != "" && string(route.Method) != method {
      continue;
    }
    if route.Pattern == nil {
      continue
    }
    matches := route.Pattern.FindStringSubmatch(path)
    if matches == nil {
      continue
    }
    pathArgs = matches[1:]
    result = route
    return
  }
  return
}

func (router *Router) Add(route *Route) {
	router.lock.Lock()
	defer router.lock.Unlock()
	if route.Name != "" {
		e := router.routes.Front()
		for e != nil {
			r := e.Value.(*Route)
			if r.Name != "" && r.Name == route.Name {
				oldE := e
				e = e.Next()
        router.routes.Remove(oldE)
			} else {
				e = e.Next()
			}
		}
		router.nameToRoute[route.Name] = route
	}
	router.routes.PushBack(route)
}

func (router *Router) Reverse(name string) (result string) {
	router.lock.Lock()
	defer router.lock.Unlock()
	if route, ok := router.nameToRoute[name]; ok {
		result = route.Reverse
	}
	return
}

func (app *App) Get(pattern string, name string, reverse string, handler Handler) {
	app.Router.Add(&Route{
		Method:  Get,
		Name:    name,
		Reverse: reverse,
		Pattern: regexp.MustCompile(pattern),
		Handler: handler})
}

func (app *App) Post(pattern string, name string, reverse string, handler Handler) {
	app.Router.Add(&Route{
		Method:  Post,
		Name:    name,
		Reverse: reverse,
		Pattern: regexp.MustCompile(pattern),
		Handler: handler})
}


