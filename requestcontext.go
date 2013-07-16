package phinney

func NewRequestContextPlugin() (plugin *Plugin) {
  plugin = &Plugin{}
  plugin.StartRequest = func(app *App, req *Request) (halt bool, err error) {
      req.Context["App"] = app
      return
  }
  return
}
