package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestCopy(t *testing.T) {
	a, err := GetRandBytes(8)
	if err != nil {
		t.Error(err)
	}
	b, err := GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}
	c, err := GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	// dst > src
	Copy(b, a)
	if !bytes.Equal(b[:8], a) {
		t.Error("incorrect copying")
	}

	// dst < src
	Copy(b, c)
	if !bytes.Equal(b, c[:16]) {
		t.Error("incorrect copying")
	}

	// dst = src
	b2, err := GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}
	Copy(b, b2)
	if !bytes.Equal(b, b2) {
		t.Error("incorrect copying")
	}
}

func TestCompare(t *testing.T) {
	a, err := GetRandBytes(8)
	if err != nil {
		t.Error(err)
	}
	b, err := GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}
	c := make([]byte, 16)
	copy(c, b)

	// not equal
	if Equal(a, b) {
		t.Error("expected not equal")
	}

	// equal
	if !Equal(b, c) {
		t.Error("expected equal")
	}
}

func TestGetRandBytes(t *testing.T) {
	b, err := GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	if bytes.Equal(b, make([]byte, 32)) {
		t.Error("bytes not random")
	}
}

func TestHash(t *testing.T) {
	known := make(map[string]string)
	known[""] = "DldRwCblQ7Loqy6wYJnaodHl30d3j3eH+qtFzfEv46g="
	known["hash"] = "l+2qaVlkOBNtzRKFU+kEvAP1JkJvcn0nC2mEH7bPUNM="
	known["test"] = "kosgNmlD4q/RHrwOri5TqTvxd6T881vMZNUDcE5l4gI="

	for k, v := range known {
		if base64.StdEncoding.EncodeToString(Hash([]byte(k))) != v {
			t.Error("digest doesn't match known values")
		}
	}
}

func TestMemClr(t *testing.T) {
	b, err := GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	MemClr(b)
	for i := range b {
		if b[i] != 0 {
			t.Error("memclr unsuccessful")
		}
	}
}

func TestMemSet(t *testing.T) {
	b, err := GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	MemSet(b, 0xdb)
	for i := range b {
		if b[i] != 0xdb {
			t.Error("memset unsuccessful")
		}
	}
}

func TestMemScr(t *testing.T) {
	b := make([]byte, 32)

	if err := MemScr(b); err != nil {
		t.Error(err)
	}
	if bytes.Equal(b, make([]byte, 32)) {
		t.Error("memscr unsuccessful")
	}
}

func TestSealOpen(t *testing.T) {
	// Declare the plaintext and the key.
	m, err := GetRandBytes(64)
	if err != nil {
		t.Error(err)
	}
	k, err := GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	// Encrypt the message.
	x, err := Seal(m, k)
	if err != nil {
		t.Error(err)
	}

	// Decrypt the message.
	dm, err := Open(x, k)
	if err != nil {
		t.Error(err)
	}

	// Verify that the plaintexts match.
	if !bytes.Equal(m, dm) {
		t.Error("incorrect decryption")
	}

	// Generate an incorrect key.
	ik, err := GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	// Attempt decryption with the incorrect key.
	dm, err = Open(x, ik)
	if err == nil {
		t.Error("expected error with incorrect key")
	}
	if dm != nil {
		t.Error("expected nil plaintext; got", dm)
	}

	// Modify the ciphertext somewhat.
	for i := range x {
		if i%2 == 0 {
			x[i] = 0xdb
		}
	}

	// Attempt decryption of the invalid ciphertext with the correct key.
	dm, err = Open(x, k)
	if err == nil {
		t.Error("expected error with modified ciphertext")
	}
	if dm != nil {
		t.Error("expected nil plaintext; got", dm)
	}

	// Generate a key of an invalid length.
	ik, err = GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}

	// Attempt encryption with the invalid key.
	ix, err := Seal(m, ik)
	if err == nil {
		t.Error("expected error with invalid key")
	}
	if ix != nil {
		t.Error("expected nil ciphertext; got", dm)
	}

	// Attempt decryption with the invalid key.
	im, err := Open(x, ik)
	if err == nil {
		t.Error("expected error with invalid key")
	}
	if im != nil {
		t.Error("expected nil plaintext; got", dm)
	}
}
