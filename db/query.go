package db

import (
  "fmt"
  "strings"
)

type Query struct {
  Table *Table
  Parent *Query
  JoinPath []string
  Key string
  Value interface {}
  Conn *Conn
  Fill []string
}

type QueryResult struct {
  Query *Query
  Error error
}


func (t *Query) root() (result *Query) {
  for curr := t; curr != nil; curr = curr.Parent {
    if curr.Parent == nil {
      result = curr
      break
    }
  }
  if result == nil {
    panic("expecting non-nil root")
  }
  return
}

func (t *Query) SetConn(conn *Conn) {
  t.root().Conn = conn
}

func (t *Query) conn() (result *Conn) {
  root := t.root()
  result = root.Conn
  if result == nil {
    return root.Table.schema.Conn
  }
  return
}


func (t *Query) sourceTable() (result *Table) {
  return t.root().Table
}

func (t *Query) Filter(key string, value interface{}) *Query {
  keys := strings.Split(key, ".")
  path := keys[:len(keys) - 1]
  var fill []string
  if len(path) > 0 {
    fill = append(fill, strings.Join(path, "."))
  }
  return &Query{
    Parent: t,
    Fill: fill,
    JoinPath: path, Key: keys[len(keys) - 1], Value: value}
}


func NewQuery(table *Table) (result *Query) {
  return &Query{Table: table}
}


func (q *Query) compile() (result string, bind []interface{}, err error) {
  type Join struct {
    path []string
    alias string
  }

  joins := make(map[string]Join)
  for curr := q; curr.Parent != nil; curr = curr.Parent {
    if len(curr.JoinPath) == 0 {
      continue
    }
    for i, _ := range curr.JoinPath {
      j := Join{}
      j.path = curr.JoinPath[:i + 1]
      j.alias = strings.Join(j.path, "__")
      joins[j.alias] = j
      // alias is now like "foo" or "foo__bar"
    }
  }

  var sql SQL
  sql.WriteKeyword("SELECT")
  table := q.sourceTable()
  i := 0
  for col, _ := range table.Columns {
    if i > 0 {
      sql.Write(", ")
    }
    sql.WriteIdentifier(table.Name)
    sql.Write(".")
    sql.WriteIdentifier(col)
    i += 1
  }
  sql.WriteKeyword("FROM")
  sql.WriteIdentifier(table.Name)
  sql.WriteKeyword("AS")
  sql.WriteIdentifier(table.Name)

  for toAlias, join := range joins {
    sql.WriteKeyword("INNER JOIN")
    to := join.path[len(join.path) - 1]
    sql.WriteIdentifier(to)
    sql.Write(" ")
    sql.WriteIdentifier(toAlias)
    sql.WriteKeyword("ON")
    from := table
    if len(join.path) > 1 {
      from = table.schema.Table(join.path[len(join.path) - 2])
    }
    fk, exists := from.ForeignKeys[join.path[len(join.path) - 1]]
    if !exists {
      err = fmt.Errorf("cant find foreign key")
      return
    }
    sql.WriteIdentifier(toAlias)
    sql.Write(".")
    sql.WriteIdentifier(fk.ToCol)
    sql.WriteKeyword("=")
    fromAlias := table.Name
    if len(join.path) > 1 {
      fromAlias = strings.Join(join.path[:len(join.path) - 1], "__")
    }
    sql.WriteIdentifier(fromAlias)
    sql.Write(".")
    sql.WriteIdentifier(fk.FromCol)
  }

  clauseNum := 0
  for curr := q; curr.Parent != nil; curr = curr.Parent {
    if curr.Key == "" {
      continue
    }
    if clauseNum == 0 {
      sql.WriteKeyword("WHERE")
    } else {
      sql.WriteKeyword("AND")
    }
    clauseNum += 1
    if len(curr.JoinPath) > 0 {
      sql.WriteIdentifier(strings.Join(curr.JoinPath, "__"))
      sql.Write(".")
    }
    sql.WriteIdentifier(curr.Key)
    sql.WriteEq(curr.Value)
  }
  result = sql.String()
  bind = sql.Bind()
  return
}

func (q *Query) Each(each func (row *Row) (stop bool)) (err error) {
  rootSQL, rootBind, err := q.compile()
  if err != nil {
    return err
  }
  sourceTable := q.sourceTable()
  conn := q.conn()
  rowDicts, err := conn.Query(rootSQL, rootBind)
  rows := make([]*Row, 0, len(rowDicts))
  for _, rowDict := range rowDicts {
    row := &Row{RowDict: rowDict, Table: sourceTable}
    rows = append(rows, row)
  }

  for _, path := range q.fillPaths() {
    err = q.fillRows(conn, rows, path)
    if err != nil {
      return
    }
  }
  for _, row := range rows {
    stop := each(row)
    if stop {
      break
    }
  }
  return
}

func (q *Query) fillRows(conn *Conn, rows []*Row, path []string) (err error) {
  currSeq := rows
  for _, edge := range path {
    if len(currSeq) == 0 {
      return
    }
    pks := make([]interface{}, 0)
    srcTable := currSeq[0].Table
    fk, exists := srcTable.ForeignKeys[edge]
    if !exists {
      err = fmt.Errorf("expecting edge %q", edge)
      return
    }
    for _, row := range currSeq {
      val := row.Get(fk.FromCol)
      if val != nil {
        pks = append(pks, val)
      }
    }
    if len(pks) == 0 {
      return
    }
    toTable := srcTable.schema.Table(fk.ToName)
    var dstRows []*Row
    dstRows, err = toTable.Query().Filter(fk.ToCol, pks).All()
    if err != nil {
      return
    }
    keyToDstRows := make(map[interface{}][]*Row)
    for _, dstRow := range dstRows {
      key := dstRow.Get(fk.ToCol)
      seq := keyToDstRows[key]
      seq = append(seq, dstRow)
      keyToDstRows[key] = seq
    }
    var nextSeq []*Row
    for _, row := range currSeq {
      fromKey := row.Get(fk.FromCol)
      if fromKey == nil {
        continue
      }
      vals, exists := keyToDstRows[fromKey]
      if !exists {
        continue
      }
      if fk.Type == HasOne {
        row.Set(edge, vals[0])
        nextSeq = append(nextSeq, vals[0])
      } else {
        row.Set(edge, vals)
        for _, v := range vals {
          nextSeq = append(nextSeq, v)
        }
      }
    }
    currSeq = nextSeq
  }
  return
}

func (q *Query) fillPaths() (result [][]string) {
  result = make([][]string, 0)
  for curr := q; curr != nil; curr = curr.Parent {
    if curr.Fill != nil {
      result = append(result, curr.Fill)
    }
  }
  return
}

func (q *Query) One() (result *Row, err error) {
 err = q.Each(func(row *Row) (stop bool) {
    result = row
    stop = true
    return
  })
  return
}

func (q *Query) All() (rows []*Row, err error) {
  rows = make([]*Row, 0)
  err = q.Each(func(row *Row) (stop bool) {
    rows = append(rows, row)
    stop = false
    return
  })
  return
}

func (q *Query) FillRelated(path ...string) (result *Query) {
  result = &Query{
    Fill: path,
    Parent: q,
  }
  return
}

