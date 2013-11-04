package web

import (
  . "launchpad.net/gocheck"
  "bytes"
  "testing"
  _ "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "os"
)

func (s *MySuite) Testquery(c *C) {
}


type DummyResponseWriter struct {
  headers http.Header
  *bytes.Buffer
  status int
}

func newDummyResponseWriter() *DummyResponseWriter {
  return &DummyResponseWriter{
    headers: make(http.Header),
    Buffer: &bytes.Buffer{},
  }
}

func (d *DummyResponseWriter) Header() http.Header {
  return d.headers
}

func (d *DummyResponseWriter) WriteHeader(status int) {
  d.status = status
}

type testHandler struct {
  Handler
}

func (h *testHandler) Get() (err error) {
  return WriteHTML(h, "hello")
}

type templateHandler struct {
  Handler
}

func (t *templateHandler) Get() (err error) {
  p := t.Request().URL.Query().Get("t")
  err = SendTemplate(t, p, map[string]string{})
  return
}

func (s *MySuite) TestTemplate(c *C) {
  f, err := ioutil.TempFile("", "template")
  f.Write([]byte(`hello`))
  f.Close()
  defer os.Remove(f.Name())
  request, err := http.NewRequest("GET", "/?t=" + f.Name(), &bytes.Buffer{})
  c.Assert(err, IsNil)
  c.Assert(request, Not(IsNil))

  app := &App{}
  app.Route("/", &templateHandler{})
  responseWriter := newDummyResponseWriter()
  app.ServeHTTP(responseWriter, request)
  c.Assert(responseWriter.status, Equals, 200)
  c.Assert(responseWriter.String(), Equals, "hello")
}


func (s *MySuite) TestApp(c *C) {
  request, err := http.NewRequest("GET", "/", &bytes.Buffer{})
  c.Assert(err, IsNil)
  c.Assert(request, Not(IsNil))
  responseWriter := newDummyResponseWriter()
  app := &App{}
  app.Route("/", &testHandler{})
  app.ServeHTTP(responseWriter, request)
  c.Assert(responseWriter.status, Equals, 200)
  c.Assert(responseWriter.String(), Equals, "hello")
  request, err = http.NewRequest("GET", "/foo", &bytes.Buffer{})
  responseWriter = newDummyResponseWriter()
  app.ServeHTTP(responseWriter, request)
  c.Assert(responseWriter.status, Equals, 404)
}

type testJSONHandler struct {
  Handler
}

func (h *testJSONHandler) Post() (err error) {
  var data testJSONData
  err = ReadJSON(h, &data)
  if err != nil {
    return
  }
  err = WriteJSON(h, data)
  return
}

type testJSONData struct {
  Name string
  Number int
}

func (s *MySuite) TestJSON(c *C) {
  buf := &bytes.Buffer{}
  var data testJSONData
  data.Name = "Brandon"
  data.Number = 5
  err := json.NewEncoder(buf).Encode(data)
  c.Assert(err, IsNil)
  request, err := http.NewRequest("POST", "/", buf)
  c.Assert(err, IsNil)
  request.Header.Add("Content-Type", "application/json")

  c.Assert(err, IsNil)
  responseWriter := newDummyResponseWriter()
  app := &App{}
  h := &testJSONHandler{}
  app.Route("/", h)
  app.ServeHTTP(responseWriter, request)
  c.Assert(responseWriter.status, Equals, 200)
  c.Assert(responseWriter.headers.Get("Content-Type"), Equals, "application/json")
  var result testJSONData
  err = json.NewDecoder(responseWriter).Decode(&result)
  c.Assert(err, IsNil)
  c.Assert(data, Equals, result)


}

type MySuite struct{}
var _ = Suite(&MySuite{})

// Hook up gocheck into the "go test" runner.
func TestDB(t *testing.T) { TestingT(t) }

