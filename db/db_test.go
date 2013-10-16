package db

import (
  . "launchpad.net/gocheck"
  "testing"
  "os"
  "fmt"
)

var driver = os.Getenv("DB_ENGINE")
var uri = os.Getenv("DB_URL")

var schema = NewSchema()

func (s *MySuite) Testquery(c *C) {
  conn := NewConn(driver, uri)

  result, err := conn.Query("SELECT 1 as foo", nil)
  c.Assert(err, IsNil)
  c.Check(len(result), Equals, 1)
  c.Check(result[0].Get("foo"), Equals, int64(1))

  _, err = conn.Exec(`create table bar2 (id int primary key, a text)`, nil)
  c.Assert(err, IsNil)
  _, err = conn.Exec(`insert into bar2 (id, a) values (1, 'hi')`, nil)
  c.Assert(err, IsNil)
  barRows, err := conn.Query(`select * from bar2`, nil)
  c.Assert(err, IsNil)
  fmt.Println("bar rows: ", barRows)
  c.Assert(len(barRows), Equals, 1)

}

func (s *MySuite) TestJSON(c *C) {

  row := make(RowDict)
  row["hello"] = "there"
  result, err := row.ToJSON()
  c.Assert(err, IsNil)
  c.Check(string(result), Equals, `{"hello":"there"}`)

}

func (s *MySuite) TestDB(c *C) {

  conn := NewConn(driver, uri)
  schema.Conn = conn
  _, err := conn.Exec(`
    create table foo (
      id serial primary key,
      text_col text not null,
      int_col int not null)
    `, nil)
  c.Assert(err, IsNil)
  _, err = conn.Exec(`
    create table bar (
      id serial primary key,
      foo_id int references foo(id)
    )
    `,
    nil)
  c.Assert(err, IsNil)
  schema.Reflect()
  fooTable, exists := schema.Tables["foo"]
  c.Assert(exists, Equals, true)
  foo := fooTable.NewRow()
  foo.Set("text_col", "Hello")
  foo.Set("int_col", 25)
  c.Check(25, Equals, foo.Get("int_col"))
  c.Check("Hello", Equals, foo.Get("text_col"))
  err = foo.Save()
  c.Check(err, IsNil)

  c.Assert(foo.Int("id"), Equals, int64(1))
  row, err := fooTable.Query().Filter("id", 1).One()
  c.Assert(err, IsNil)
  c.Assert(row, Not(IsNil))
  c.Check(row.String("text_col"), Equals, "Hello")
  c.Check(row.Int("int_col"), Equals, int64(25))

  foo.Set("int_col", 36)
  err = foo.Save()
  c.Assert(err, IsNil)
  row, err = fooTable.Query().Filter("id", 1).One()
  c.Assert(err, IsNil)
  c.Assert(row, Not(IsNil))
  c.Check(row.Int("int_col"), Equals, int64(36))
  bars := schema.Tables["bar"]
  c.Assert(bars.ForeignKeys["foo"], Not(IsNil))

  bar := bars.NewRow()
  bar.Set("foo_id", foo.Get("id"))
  err = bar.Save()
  c.Assert(err, IsNil)

  row, err = bars.Query().FillRelated("foo").One()
  c.Assert(err, IsNil)
  c.Assert(row, Not(IsNil))
  aFoo := row.Row("foo")
  c.Assert(aFoo, Not(IsNil))
  c.Assert(aFoo.Int("int_col"), Equals, int64(36))

  foo, err = fooTable.Query().Filter("id", 1).FillRelated("bar").One()
  c.Assert(err, IsNil)
  c.Assert(foo.Int("id"), Equals, int64(1))
  c.Assert(len(foo.Rows("bar")), Equals, 1)
}


type MySuite struct{}
var _ = Suite(&MySuite{})

// Hook up gocheck into the "go test" runner.
func TestDB(t *testing.T) { TestingT(t) }

