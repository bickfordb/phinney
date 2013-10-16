package db

import (
  "database/sql"
  _ "github.com/lib/pq"
  "fmt"
)


type Conn struct {
  Driver string
  URI string
}

func NewConn(driver string, uri string) *Conn {
  return &Conn{
    Driver: driver,
    URI: uri,
  }
}


func (c *Conn) Exec(sql string, bind []interface{}) (result sql.Result, err error) {
  db, err := c.open()
  if err != nil {
    return
  }
  defer db.Close()
  if bind == nil {
    bind = make([]interface{}, 0)
  }
  result, err = db.Exec(sql, bind...)
  return
}

type RowDict map[string]interface{}

func NewRowDict() (result RowDict) {
  return make(RowDict)
}

func (c *Conn) open() (db *sql.DB, err error) {
  db, err = sql.Open(c.Driver, c.URI)
  return
}

func (c *Conn) QueryOne(sql string, bind[]interface{}) (row RowDict, err error) {
  rows, err := c.Query(sql, bind)
  if err != nil {
    return
  }
  if len(rows) > 0 {
    row = rows[0]
  }
  return
}

func (c *Conn) Query(sql string, bind []interface{}) (rows []RowDict, err error) {
  rows = make([]RowDict, 0)
  err = c.QueryEach(sql, bind, func (row RowDict) (stop bool) {
    rows = append(rows, row)
    return
  })
  return
}

type EachRowDict func (rowDict RowDict) (stop bool)

func (c *Conn) QueryEach(sql string, bind []interface{}, each EachRowDict) (err error) {
  fmt.Printf("query: %q, bind: %+v\n", sql, bind)
  db, err := c.open()
  if err != nil {
    return
  }
  defer db.Close()
  if bind == nil {
    bind = []interface{}{}
  }
  rows, err := db.Query(sql, bind...)
  if err != nil || rows == nil {
    return
  }
  defer rows.Close()
  var cols []string
  for rows.Err() == nil && rows.Next() {
    if cols == nil {
      cols, err = rows.Columns()
      if err != nil {
        return
      }
    }

    r := NewRowDict()
    vals := make([]interface{}, 0)
    for _, _ = range cols {
      var x interface{}
      vals = append(vals, &x)
    }
    err = rows.Scan(vals...)
    if err != nil {
      return
    }
    for idx, col := range cols {
      var val interface {} = *(vals[idx].(* interface{}))
      bs, is := val.([]uint8)
      if is {
        // try to convert strings:
        val = string(bs)
      }
      r[col] = val
    }
    if each(r) {
      break
    }
  }
  err = rows.Err()
  return
}
