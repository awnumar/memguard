package core

import "testing"

func TestError(t *testing.T) {
	for i := range errors {
		if errors[i].s != errors[i].Error() {
			t.Error("error string does not match")
		}
	}
}

func TestIsXError(t *testing.T) {
	for i := range errors {
		if !isXError(errors[i], errors[i].c) {
			t.Error("error should match itself")
		}
	}
}
