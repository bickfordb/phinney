package phinney

import (
	"fmt"
	. "launchpad.net/gocheck"
	"math"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestSnakeCase(c *C) {
	c.Check(capWordsToSnakeCase("FooBar"), Equals, "foo_bar")
	c.Check(capWordsToSnakeCase("FRooBar"), Equals, "froo_bar")
	c.Check(capWordsToSnakeCase("ID"), Equals, "id")
	c.Check(capWordsToSnakeCase("network"), Equals, "network")
}

func (s *MySuite) TestUUID(c *C) {
	conn, err := DBConn()
	c.Assert(err, IsNil)
	_, err = conn.Exec(`create extension if not exists "uuid-ossp"`)
	c.Assert(err, IsNil)
	_, err = conn.Exec(`create table uuidfoo (id uuid primary key, t text)`)
	c.Assert(err, IsNil)

	type UUIDObj struct {
		ID string
		T  string
	}
	var uuidExample UUIDObj
	var uuidExample2 UUIDObj

	uuidExample.ID = GenUUID()
	uuidExample.T = "Hello"

	err = Insert(nil, "uuidfoo", &uuidExample)
	c.Assert(err, IsNil)

	exists, err := Select(conn, "uuidfoo", &uuidExample2).Filter("ID", uuidExample.ID).One()
	c.Assert(exists, Equals, true)
	c.Assert(err, IsNil)
	c.Assert(uuidExample2.T, Equals, "Hello")
	c.Assert(uuidExample.ID, Equals, uuidExample2.ID)
}

func (s *MySuite) TestUpdate(c *C) {
	type Foo struct {
		ID        int64     `quux:"true" db-is-serial:"true"`
		CreatedAt time.Time `db-on-insert:"now()"`
		UpdatedAt time.Time `db-on-update:"now()"`
		Name      string
	}

	conn, err := DBConn()
	c.Assert(err, IsNil)
	c.Assert(conn, NotNil)
	defer conn.Close()
	conn.Exec("drop table foo")
	_, err = conn.Exec(`create table foo (id serial primary key, created_at timestamp, updated_at timestamp, name text)`)
	c.Assert(err, IsNil)

	var foo Foo
	foo.Name = "brandon"
	Insert(conn, "foo", &foo)
	c.Assert(foo.ID, Equals, int64(1))

	var foo2 Foo
	exists, err := Select(conn, "foo", &foo2).Filter("ID", 1).One()
	c.Assert(err, IsNil)
	c.Assert(exists, Equals, true)
	c.Check(foo2.Name, Equals, "brandon")

	foo.Name = "Person"
	err = Update(conn, "foo", &foo)
	c.Assert(err, IsNil)

	exists, err = Select(conn, "foo", &foo2).Filter("ID", 1).One()
	c.Assert(exists, Equals, true)
	c.Assert(err, IsNil)
	c.Check(foo2.Name, Equals, "Person")

}

func (s *MySuite) TestSelect(c *C) {
	type Quux struct {
		ID        int64
		IntVal    int64
		StringVal string
		FloatVal  float64
		TimeVal   time.Time
		BoolVal   bool
		BytesVal  []byte
	}
	var quux0 Quux
	quux0.StringVal = "hi"
	quux0.IntVal = 42
	quux0.FloatVal = 3.14
	quux0.TimeVal = time.Now().UTC()
	quux0.BoolVal = true
	quux0.BytesVal = []byte{36, 44, 38}
	conn, err := DBConn()
	c.Assert(err, IsNil)
	conn.Exec("drop table quux")
	_, err = conn.Exec(`create table quux (
    id serial,
    int_val bigint,
    string_val text,
    float_val real,
    time_val timestamp,
    bool_val bool,
    bytes_val bytea)`)
	c.Assert(err, IsNil)
	_, err = conn.Exec(`
    insert into quux
      (int_val, string_val, float_val, time_val, bool_val, bytes_val)
    values
      ($1, $2, $3, $4, $5, $6)
    `,
		quux0.IntVal,
		quux0.StringVal,
		quux0.FloatVal,
		quux0.TimeVal,
		quux0.BoolVal,
		quux0.BytesVal)
	c.Assert(err, IsNil)
	var quux1 Quux
	query := Select(conn, "quux", &quux1).Filter("ID", 1)
	sql, bind := query.Compile()
	c.Assert(sql, Equals, `SELECT "id", "int_val", "string_val", "float_val", "time_val", "bool_val", "bytes_val" FROM "quux" WHERE "id" = $1`)
	c.Assert(len(bind), Equals, 1)
	c.Assert(bind[0], Equals, 1)
	result, err := Select(conn, "quux", &quux1).Filter("ID", 1).exec(conn)
	c.Assert(err, IsNil)
	c.Assert(result, NotNil)
	exists, err := result.Next()
	c.Assert(err, IsNil)
	c.Assert(exists, Equals, true)
	c.Check(quux1.ID, Equals, int64(1))
	c.Check(quux1.IntVal, Equals, int64(42))
	c.Check(quux1.StringVal, Equals, "hi")
	if math.Abs(quux0.FloatVal-quux1.FloatVal) > 0.001 {
		c.ExpectFailure(fmt.Sprintf("expecting %v == %v", quux0.FloatVal, quux1.FloatVal))
	}

	c.Check(quux1.TimeVal.Unix(), Equals, quux0.TimeVal.Unix())
	c.Check(len(quux1.BytesVal), Equals, 3)
	c.Check(quux1.BytesVal[0], Equals, quux0.BytesVal[0])
	c.Check(quux1.BytesVal[1], Equals, quux0.BytesVal[1])
	c.Check(quux1.BytesVal[2], Equals, quux0.BytesVal[2])

	_, err = conn.Exec("drop table quux")
	c.Assert(err, IsNil)
}
