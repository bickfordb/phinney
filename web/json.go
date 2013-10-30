package web

import (
  "encoding/json"
  "net/http"
  "regexp"
)


var jsonContentType *regexp.Regexp = regexp.MustCompile(`^application[/]json.*$`)

func ReadJSON(request *http.Request, object interface{}) (err error) {
	if request.Method == "GET" {
		data := request.URL.Query().Get("params")
		if data != "" {
			err = json.Unmarshal([]byte(data), object)
		}
	} else {
    contentType := request.Header.Get("Content-Type")
		if !jsonContentType.MatchString(contentType) {
      println("is not json")
			return
		}
    println("is json")
		dec := json.NewDecoder(request.Body)
		err = dec.Decode(object)
	}
  return
}

func (h *Handler) ReadJSON(object interface{}) (err error) {
  err = ReadJSON(h.Request, object)
  return
}

func (h *Handler) WriteJSON(record interface{}) (err error) {
  h.Header().Set("Content-Type", "application/json")
  err = json.NewEncoder(h).Encode(record)
	return
}
