package db

import (
  "bytes"
  "encoding/json"
  "fmt"
  "strings"
  "regexp"
  "time"
)

type Row struct {
  RowDict
  Table *Table
}

func (r RowDict) Row(key string) (result *Row) {
  val, exists := r[key]
  if exists {
    switch t := val.(type) {
    case []*Row:
      if len(t) > 0 {
        result = t[0]
      }
    case *Row:
      result = t
    }
  }
  return
}

func (r RowDict) Rows(key string) (result []*Row) {
  val, exists := r[key]
  if exists {
    result, _ = val.([]*Row)
  }
  return
}

func (r RowDict) Get(key string) (result interface{}) {
  result, exists := r[key]
  if !exists {
    result = nil
  }
  return
}

func (r RowDict) Int(key string) (result int64) {
  val, exists := r[key]
  if exists {
    result, _ = val.(int64)
  }
  return
}


func (r RowDict) Set(key string, value interface{}) {
  r[key] = value
}

func (r *Row) check() (err error) {
  if len(r.Table.Keys) == 0 {
    err = &RowError{
      Row: r,
      Err: fmt.Errorf("expecting keys"),
    }
    return
  }
  return
}

func (r *Row) Save() (err error) {
  isNew := false
  err = r.check()
  if err != nil {
    return
  }
  for _, key := range r.Table.Keys {
    if keyVal, exists := r.RowDict[key]; !exists || keyVal == nil {
      isNew = true
    }
  }
  if isNew {
    err = r.doInsert()
  } else {
    err = r.doUpdate()
  }
  if err != nil {
    return
  }
  return
}

func (r *Row) doInsert() (err error) {
  sql, bind, err := r.compileInsert()
  if err != nil {
    err = &RowError{Row: r, Err: err}
    return
  }
  conn := r.getConn()
  returning, err := conn.QueryOne(sql, bind)
  if err != nil {
    return
  }
  for key, val := range returning {
    r.Set(key, val)
  }
  if err != nil {
    err = &RowError{Row: r, Err: err}
    return
  }
  return
}

// Do we need this?
func (r *Row) getConn() (conn *Conn) {
  conn = r.Table.schema.Conn
  return
}

type RowError struct {
  Row *Row
  Err error
}

func (r *RowError) Error() string {
  return fmt.Sprintf("row error -- row: %+v, error: %q", r.Row, r.Err.Error())
}

func (r *Row) getInsertVals() (result map[string]interface{}, err error) {
  result = make(map[string]interface{})
  for name, def := range r.Table.Columns {
    val, exists := r.RowDict[name]


    var f interface{} = def.OnInsert
    if f == nil {
      f = def.OnUpdate
    }
    if f != nil {
      switch t := f.(type) {
      case func() interface{}:
        val = t()
      case func() time.Time:
        val = t()
      case func() string:
        val = t()
      default:
        err = fmt.Errorf("unexpected insert type %+v", f)
        return
      }
      exists = true
    }

    if def.Required && !exists {
      err = fmt.Errorf("%q is required", name)
      return
    }

    if exists {
      result[name] = val
      continue
    }
  }
  return
}



func (r *Row) compileInsert() (sql string, bind []interface{}, err error) {
  bind = make([]interface{}, 0)
  keysPart := make([]string, 0)
  valsPart := make([]string, 0)
  insertVals, err := r.getInsertVals()
  if err != nil {
    return
  }
  for k, v := range insertVals {
    keysPart = append(keysPart, fmt.Sprintf(`%q`, k))
    bind = append(bind, v)
    valsPart = append(valsPart, fmt.Sprintf("$%d", len(bind)))
  }
  sql = fmt.Sprintf(
    `insert into %q (%s) values (%s) returning *`,
    r.Table.Name,
    strings.Join(keysPart, ", "),
    strings.Join(valsPart, ", "))
  return
}

func (r *Row) doUpdate() (err error) {
  bind := make([]interface{}, 0)
  sql := "UPDATE "
  sql += fmt.Sprintf("%q SET ", r.Table.Name)
  i := 0
  for col, opts := range r.Table.Columns {
    val := r.RowDict[col]
    if i > 0 {
      sql += ", "
    }
    if opts.OnUpdate != nil {
      switch t := opts.OnUpdate.(type) {
      case func() interface{}:
        val = t()
      default:
        err = fmt.Errorf("dont know how to convert: %+v", opts.OnUpdate)
      }
    }
    if val != nil {
      bind = append(bind, val)
      sql += fmt.Sprintf("%q = $%d", col, len(bind))
    } else {
      sql += fmt.Sprintf("%q = NULL", col)
    }
    i += 1
  }
  sql += " WHERE "
  wherePart := make([]string, 0)
  for _, key := range r.Table.Keys {
    val := r.Get(key)
    if val != nil {
      bind = append(bind, val)
      wherePart = append(wherePart, fmt.Sprintf("%q = $%d", key, len(bind)))
    } else {
      wherePart = append(wherePart, fmt.Sprintf("%q IS NULL", key))
    }
  }
  sql += strings.Join(wherePart, " AND ")
  sql += " RETURNING *"
  conn, err := r.Table.conn()
  if err != nil {
    return
  }
  returning, err := conn.QueryOne(sql, bind)
  if err != nil {
    return
  }
  for key, val := range returning {
    r.Set(key, val)
  }
  return
}

var notImplemented = fmt.Errorf("not implemented")

func (r RowDict) String(key string) (result string) {
  val, exists := r[key]
  if exists {
    result, _ = val.(string)
  }
  return
}

type RowSeq []*Row

var TLAs map[string]bool = map[string]bool{
  "url": true,
  "id": true,
}

func snakeCaseToCamelCase(s string) string {
  b := &bytes.Buffer{}
  parts := strings.Split(s, "_")
  for i, p := range parts {
    if i == 0 {
      b.WriteString(strings.ToLower(p))
    } else {
      if TLAs[p] {
        b.WriteString(strings.ToUpper(p))
      } else {
        b.WriteString(strings.Title(p))
      }
    }
  }
  return b.String()
}

func (r RowDict) ToJSON() (result []byte, err error) {
  v := r.ToJSONDict()
  result, err = json.Marshal(v)
  return
}

func (r RowDict) ToJSONDict() (result map[string]interface{}) {
  result = make(map[string]interface{})
  for key, val := range r {
    key := snakeCaseToCamelCase(key)
    switch val.(type) {
    case RowDict:
      val = val.(RowDict).ToJSONDict()
    case *Row:
      val = val.(* Row).ToJSONDict()

    }
    result[key] = val
  }
  return
}

func (r *Row) MarshalJSON() (result []byte, err error) {
  result, err = json.Marshal(r.ToJSONDict())
  return
}

func (r *Row) UnmarshalJSON(bs []byte) (err error) {
  var data map[string]interface{} = make(map[string]interface{})
  err = json.Unmarshal(bs, &data)
  if err != nil {
    return
  }
  r.fromJSON(data)
  return
}

var camelCasePattern = regexp.MustCompile(`[a-z][a-z0-9-]*|[A-Z][A-Z]*[a-z]*`)

func camelCaseToSnakeCase(src string) (result string) {
  for _, p := range camelCasePattern.FindAllString(src, -1) {
    p = strings.ToLower(p)
    if result == "" {
      result += p
    } else {
      result += "_" + p
    }
  }
  return
}

func (r *Row) fromJSON(data map[string]interface{}) {
  for key, val := range data {
    key = camelCaseToSnakeCase(key)
    _, isKey := r.Table.ForeignKeys[key]
    switch t := val.(type) {
    case map[string]interface{}:
      if !isKey {
        continue
      }
      row := r.Table.schema.Table(key).NewRow()
      row.fromJSON(t)
      r.Set(key, row)
    case []interface{}:
      if !isKey {
        continue
      }
      var rows []*Row
      for _, i := range t {
        switch t0 := i.(type) {
        case map[string]interface{}:
            row := r.Table.schema.Table(key).NewRow()
            row.fromJSON(t0)
            rows = append(rows, row)
        }
      }
      if rows != nil {
        r.Set(key, rows)
      }
    case string, bool, float64, nil:
      r.Set(key, val)
    }
  }
  return
}

