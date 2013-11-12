package db

import (
  "fmt"
  "regexp"
  "os"
)

var nextValPat *regexp.Regexp = regexp.MustCompile(`^nextval[(].*[)]`)

type Schema struct {
  tables map[string]*Table
  Conn *Conn
}

func NewSchema() *Schema {
  driver := os.Getenv("DB_ENGINE")
  uri := os.Getenv("DB_URL")

  return &Schema{
    tables: make(map[string]*Table),
    Conn: NewConn(driver, uri),
  }
}

func (schema *Schema) Table(name string) (result *Table) {
  result, exists := schema.tables[name]
  if !exists {
    result = &Table{}
    result.Name = name
    result.schema = schema
    result.Columns = make(map[string]*ColumnOptions)
    result.ForeignKeys = make(map[string]*ForeignKey)
    schema.tables[name] = result
  }
  return
}

func (schema *Schema) Reflect() (err error) {
  if schema.Conn == nil {
    err = schema.toError(fmt.Errorf("expecting conn"))
    return
  }
  conn := schema.Conn
  rows, err := conn.Query(`
    select * from
    information_schema.tables where table_type = 'BASE TABLE' and table_schema='public'`, nil)
  if err != nil {
    return
  }
  //if true {
  //  return
  //}
  for _, row := range rows {
    table := row["table_name"].(string)
    t := schema.Table(table)
    colRows, err := conn.Query(`
      select *
      from information_schema.columns
      where table_name = $1`, []interface{}{table})
    if err != nil {
      return err
    }
    for _, colRow := range colRows {
      columnName := colRow["column_name"].(string)
      isNullable := colRow["is_nullable"].(string)
      dataType := colRow["data_type"].(string)
      columnName = columnName
      isNullable = isNullable
      dataType = dataType
      colDef := t.Column(columnName)
      colDef.IsNullable = isNullable == "YES"
      // by default required == not null

      colDef.Required = !colDef.IsNullable
      def, ok := colRow["column_default"].(string)
      if ok && nextValPat.MatchString(def)  {
        colDef.Required = false
        colDef.IsSerial = true
      }
    }
    keyRows, err := conn.Query(`
      select
      *
      from information_schema.key_column_usage
      where table_name = $1
      `, []interface{}{table})
    if err != nil {
      return err
    }
    for _, keyRow := range keyRows {
      col, ok := keyRow["column_name"].(string)
      if ok {
        t.Keys = append(t.Keys, col)
      }
    }
    // reflect foreign keys


  }
  var fks []RowDict
  fks, err = conn.Query(`
    SELECT
      tc.constraint_name,
      tc.table_name,
      kcu.column_name,
      ccu.table_name AS foreign_table_name,
      ccu.column_name AS foreign_column_name
    FROM
      information_schema.table_constraints AS tc
      JOIN information_schema.key_column_usage AS kcu
        ON tc.constraint_name = kcu.constraint_name
      JOIN information_schema.constraint_column_usage AS ccu
        ON ccu.constraint_name = tc.constraint_name
      WHERE constraint_type = 'FOREIGN KEY'`,
      nil)
  for _, fk := range fks {
    dst := schema.Table(fk.String("foreign_table_name"))
    src := schema.Table(fk.String("table_name"))

    toF := &ForeignKey{
      ToName: dst.Name,
      ToCol: fk.String("foreign_column_name"),
      FromCol: fk.String("column_name"),
      FromName: src.Name,
      Type: HasOne,
    }
    fromF := &ForeignKey{
      ToName: toF.FromName,
      ToCol: toF.FromCol,
      FromName: toF.ToName,
      FromCol: toF.ToCol,
      Type: HasMany,
    }
    dst.ForeignKeys[src.Name] = fromF
    src.ForeignKeys[dst.Name] = toF
  }
  return
}


type SchemaError struct {
  Schema *Schema
  Err error
}

func (schema *Schema) toError(err error) (result error) {
  if err != nil {
    result = &SchemaError{
      Err: err,
      Schema: schema,
    }
  }
  return
}

func (e *SchemaError) Error() string {
  return fmt.Sprintf("SchemaError{Schema: %+v, Err: %q}", e.Schema, e.Err.Error())
}

