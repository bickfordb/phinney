package phinney

import (
)

func (p *Project) runWeb() (procExit chan error, stopProc chan error) {
  procExit = make(chan error, 1)
  stopProc = make(chan error, 1)
  go func() {
    defer close(procExit)
    var err error
    proc, err := p.Exec(p.BuildDir() + "/web")
    if err != nil {
      procExit <- err
      return
    }
    go func() {
      select {
      case <-stopProc:
        if proc != nil {
          proc.Kill()
          proc.Wait()
        }
      }
    }()
    if proc != nil {
      _, err = proc.Wait()
    }
    procExit <- err
  }()
  return
}

