package phinney


import (
  "flag"
  "fmt"
)

func web() (err error) {
  project, err := NewProject(".")
  firstRun := false
  if err != nil { return }
  for {
    log.Debug("loop")
    buildErr := project.buildWeb()
    if buildErr != nil {
      log.Debug("build error: %s", buildErr)
      if firstRun {
        err = buildErr
        return
      }
    }
    procDone, stopProc := project.runWeb()
    firstRun = false
    defer close(stopProc)
    log.Debug("wait")
    changes := project.watch()
    select {
    case watchErr := <-changes:
      if watchErr != nil {
        log.Debug("watch error: %s", watchErr.Error())
      }
      stopProc <- nil
      select {
      case procErr := <-procDone:
        log.Debug("process done")
        if procErr != nil {
          log.Debug("process error: %s", procErr.Error())
        }
      }
    }
  }
  return
}

func main() (err error) {
  flag.Parse()
  switch cmd := flag.Arg(0); cmd {
  case "web":
    err = web()
  case "":
    err = fmt.Errorf("expecting a command")
  default:
    err = fmt.Errorf("unexpected command %q", cmd)
  }
  return
}

func Main() {
  err := main()
  if err != nil {
    log.Error("unexpected error: %+v", err.Error())
  }
  return
}
