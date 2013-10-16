package db

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

func (t *Table) DefineColumn(name string, options ColumnOptions) {
}

func (t *Table) NewRow() *Row {
  row := &Row{}
  row.Table = t
  row.RowDict = make(RowDict)
  return row
}

type Validator func(interface {}) error

type ColumnOptions struct {
  OnInsert func() interface{}
  OnUpdate func() interface{}
  Validator Validator
  IsNullable bool
  IsSerial bool
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

