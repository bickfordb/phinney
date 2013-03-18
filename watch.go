package phinney

import (
  "os"
  "path/filepath"
  "github.com/howeyc/fsnotify"
)

func (project *Project) watch() (changes chan error) {
  changes = make(chan error, 1)
  go func() {
    defer close(changes)
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
      changes <- err
      return
    }
    defer watcher.Close()
    err = watcher.Watch(project.SrcDir())
    if err != nil {
      changes <- err
      return
    }
    err = filepath.Walk(project.SrcDir(), func(path string, info os.FileInfo, walkErr error) (err error) {
      if info.IsDir() {
        err = watcher.Watch(path)
      }
      return
    })
    if err != nil {
      changes <- err
      return
    }
    select {
    case <-watcher.Event:
      changes <- nil
    case e := <-watcher.Error:
      changes <- e
    }
  }()
  return
}

