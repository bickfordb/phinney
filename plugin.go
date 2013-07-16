package phinney

type Plugin struct {
  StartRequest func(app *App, req *Request) (halt bool, err error)
  BeforeHandler func(app *App, req *Request) (halt bool, err error)
  AfterResponse func(app *App, req *Request) (err error)
}



