package phinney

type Plugin struct {
  StartRequest func(app *App, req *Request) (err error)
}



