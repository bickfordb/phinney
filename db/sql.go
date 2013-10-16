package db

import (
  "bytes"
  "fmt"
  "strings"
)

type SQL struct {
  buf bytes.Buffer
  bind []interface{}
}

func NewSQL() *SQL {
  return &SQL{}
}

func (s *SQL) String() string {
  return s.buf.String()
}

func (s *SQL) Bind() []interface{} {
  return s.bind
}

func (s *SQL) BindValue(value interface{}) {
  if value == nil {
    s.WriteKeyword("NULL")
  } else {
    if slice, ok := value.([]interface{}); ok {
      s.WriteKeyword("(")
      for _, v := range slice {
        s.BindValue(v)
      }
      s.WriteKeyword(")")
    } else {
      s.bind = append(s.bind, value)
      s.Write(fmt.Sprintf("$%d", len(s.bind)))
    }
  }
}

func (s *SQL) Write(aString string) {
  s.buf.WriteString(aString)
}

func (s *SQL) WriteIdentifier(identifier string) {
  s.Write(fmt.Sprintf("%q", identifier))
}

func (s *SQL) WriteKeyword(kw string) {
  s.Write(" " + strings.ToUpper(kw) + " ")
}

func (s *SQL) WriteEq(val interface{}) {
  if val == nil {
    s.WriteKeyword("IS")
  } else if _, ok := val.([]interface{}); ok {
    s.WriteKeyword("IN")
  } else {
    s.WriteKeyword("=")
  }
  s.BindValue(val)
}

