package phinney

import (
)

type App struct {
  routes []*Route
}

func NewApp() *App {
  return &App{make([]*Route, 0)}
}
