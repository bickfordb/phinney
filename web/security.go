package web

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/rand"
	//"io/ioutil"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var rng *rand.Rand

func init() {
	src := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(src)
}

func to64(b []byte) string {
	buf := &bytes.Buffer{}
	enc := base64.NewEncoder(base64.URLEncoding, buf)
	enc.Write(b)
	enc.Close()
	return strings.TrimRight(buf.String(), "=")
}

func ToSaltedPassword(password string, salt string) (result string) {
	// 8 bytes of randomness
	if salt == "" {
		rnd := rng.Uint32()
		salt = to64([]byte{byte(rnd >> 24), byte(rnd >> 16), byte(rnd >> 8), byte(rnd)})
	}
	enc := hmac.New(sha256.New, []byte(password))
	lhs := enc.Sum([]byte(salt))
	return fmt.Sprintf("%s$%s", to64(lhs), salt)
}

func CheckPassword(password, salted string) bool {
	p := strings.SplitN(salted, "$", 2)
	if len(p) != 2 {
		return false
	}
	return salted == ToSaltedPassword(password, p[1])
}

func CheckString1(s string, key []byte) (result string, at time.Time) {
	sepIdx := strings.LastIndex(s, "$")
	if sepIdx < 0 {
		return
	}
	s0 := s[:sepIdx]
	buf := bytes.NewBuffer([]byte(s[sepIdx+1:]))
	dec := base64.NewDecoder(base64.URLEncoding, buf)
	var timestamp int64
	err := binary.Read(dec, binary.BigEndian, &timestamp)
	if err != nil {
		return
	}
	at0 := time.Unix(timestamp, 0)
	expected := SecureString2(s0, key, at0)
	if expected == s {
		result = s0
		at = at0
	}
	return
}

func CheckString(s string) (result string, at time.Time) {
	result, at = CheckString1(s, SecureStringSecret)
	return
}

var SecureStringSecret []byte = []byte("secret")

func SecureString(s string) string {
	return SecureString1(s, SecureStringSecret)
}

func SecureString1(s string, key []byte) string {
	return SecureString2(s, key, time.Now().UTC())
}

func SecureString2(s string, key []byte, at time.Time) string {
	buf := &bytes.Buffer{}
	buf.WriteString(s)
	h := hmac.New(sha1.New, key)
	ts := at.Unix()
	binary.Write(h, binary.BigEndian, ts)
	h.Write([]byte(s))
	sig := h.Sum(nil)
	buf.WriteString("$")
	enc := base64.NewEncoder(base64.URLEncoding, buf)
	binary.Write(enc, binary.BigEndian, ts)
	enc.Write(sig)
	enc.Close()
	return buf.String()
}

func (h *Handler) AddSecureCookie(cookie *http.Cookie) {
	cookie0 := *cookie
	cookie0.Value = SecureString(cookie0.Value)
	h.Header().Add("Set-Cookie", cookie0.String())
}

func (h *Handler) GetSecureCookie(name string) (s string, at time.Time) {
	cookie, err := h.Request.Cookie(name)
	if cookie == nil || err != nil {
		return
	}
	s, at = CheckString(cookie.Value)
	return
}

func SecureObject(o interface{}) (s string, err error) {
	bs, err := json.Marshal(o)
	if err != nil {
		return
	}
	s = SecureString(string(bs))
	return
}

func DecodeSecureObject(encoded string, objPtr interface{}) (decoded bool, at time.Time, err error) {
	s, at := CheckString(encoded)
	if s == "" {
		return
	}
	err = json.Unmarshal([]byte(s), objPtr)
	if err != nil {
		return
	}
	decoded = true
	return
}
