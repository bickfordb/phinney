package web

import (
  "encoding/json"
  "regexp"
)


var jsonContentType *regexp.Regexp = regexp.MustCompile(`^application[/]json.*$`)

func ReadJSON(handler Handler, object interface{}) (err error) {
	if handler.Request().Method == "GET" {
		data := handler.Request().URL.Query().Get("params")
		if data != "" {
			err = json.Unmarshal([]byte(data), object)
		}
	} else {
    contentType := handler.Request().Header.Get("Content-Type")
		if !jsonContentType.MatchString(contentType) {
			return
		}
		dec := json.NewDecoder(handler.Request().Body)
		err = dec.Decode(object)
	}
  return
}

func WriteJSON(h Handler, record interface{}) (err error) {
  h.Header().Set("Content-Type", "application/json")
  err = json.NewEncoder(h).Encode(record)
	return
}

func WriteJSONRPCResult(h Handler, record interface{}) (err error) {
  return WriteJSON(h, JSONObj{"result": record})
}

type JSONObj map[string]interface{}

func WriteJSONRPCError(h Handler, code int, msg string) (err error) {
  return WriteJSON(h, JSONObj{"error": JSONObj{"message": msg, "code": code}})
}

func ReadJSONRPC(h Handler, params interface{}) (err error) {
  var req struct {
    Params json.RawMessage
  }
  err = ReadJSON(h, &req)
  if err != nil {
    return
  }
  if req.Params != nil {
    err = json.Unmarshal(req.Params, params)
  }
  return
}

