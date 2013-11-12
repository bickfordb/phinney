package db

import (
	"crypto/rand"
	"fmt"
	"io"
  "time"
)

type Table struct {
  Name string
  Keys []string
  schema *Schema
  Columns map[string]*ColumnOptions
  ForeignKeys map[string]*ForeignKey
}


func (t *Table) Query() *Query {
  return NewQuery(t)
}

func (t *Table) Column(key string) (result *ColumnOptions) {
  result, exists := t.Columns[key]
  if !exists {
    result = &ColumnOptions{}
    t.Columns[key] = result
  }
  return
}


func (t *Table) NewRow() *Row {
  row := &Row{}
  row.Table = t
  row.RowDict = make(RowDict)
  return row
}

type Validator func(interface {}) error

type ColumnOptions struct {
  OnInsert interface{}
  OnUpdate interface{}
  Validator Validator
  IsNullable bool
  IsSerial bool
  IsPublic bool
  ForeignKey *ForeignKey
  Required bool
}

type ForeignKeyType string

const (
  HasMany ForeignKeyType = "has many"
  HasOne = "has one"
)

type ForeignKey struct {
  ToName string
  ToCol string
  FromCol string
  FromName string
  Type ForeignKeyType
}

func (t *Table) conn() (conn *Conn, err error) {
  conn = t.schema.Conn
  return
}

func UUID() string {
	xs := make([]byte, 16)
	io.ReadFull(rand.Reader, xs)
	return fmt.Sprintf("%x-%x-%x-%x-%x", xs[0:4], xs[4:6], xs[6:8], xs[8:10], xs[10:])
}

func At() time.Time {
  return time.Now().UTC()
}
