package web

import (
	"testing"
)

func TestSecureString(t *testing.T) {
	secret := []byte("hello")
	xs := SecureString1("hi", secret)
	result, _ := CheckString1(xs, secret)
	expected := "hi"
	if result != expected {
		t.Errorf("expecting %q but got %q (%q)", expected, result, xs)
	}
}
