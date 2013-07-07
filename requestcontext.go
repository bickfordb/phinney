package phinney

func NewRequestContextPlugin() (plugin *Plugin) {
  plugin = &Plugin{}
  plugin.StartRequest = func(app *App, req *Request) (err error) {
      req.Context["App"] = app
      return
  }
  return
}
