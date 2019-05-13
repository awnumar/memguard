package core

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestCopy(t *testing.T) {
	a := make([]byte, 8)
	Scramble(a)
	b := make([]byte, 16)
	Scramble(b)
	c := make([]byte, 32)
	Scramble(c)

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
	b2 := make([]byte, 16)
	Scramble(b2)
	Copy(b, b2)
	if !bytes.Equal(b, b2) {
		t.Error("incorrect copying")
	}
}

func TestMove(t *testing.T) {
	a := make([]byte, 32)
	Scramble(a)
	b := make([]byte, 32)
	Scramble(b)

	Move(a, b)
	if !bytes.Equal(b, make([]byte, 32)) {
		t.Error("src buffer was not wiped")
	}
}

func TestCompare(t *testing.T) {
	a := make([]byte, 8)
	Scramble(a)
	b := make([]byte, 16)
	Scramble(b)
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

	c[8] = ^c[8]

	// not equal
	if Equal(b, c) {
		t.Error("expected not equal")
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

func TestWipe(t *testing.T) {
	b := make([]byte, 32)
	Scramble(b)
	Wipe(b)
	for i := range b {
		if b[i] != 0 {
			t.Error("wipe unsuccessful")
		}
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// Declare the plaintext and the key.
	m := make([]byte, 64)
	Scramble(m)
	k := make([]byte, 32)
	Scramble(k)

	// Encrypt the message.
	x, err := Encrypt(m, k)
	if err != nil {
		t.Error("expected no errors; got", err)
	}

	// Decrypt the message.
	dm := make([]byte, len(x)-Overhead)
	length, err := Decrypt(x, k, dm)
	if err != nil {
		t.Error("expected no errors; got", err)
	}
	if length != len(x)-Overhead {
		t.Error("unexpected plaintext length; got", length)
	}

	// Verify that the plaintexts match.
	if !bytes.Equal(m, dm) {
		t.Error("decrypted plaintext does not match original")
	}

	// Attempt decryption /w buffer that is too small to hold the output.
	out := make([]byte, len(x)-Overhead-1)
	length, err = Decrypt(x, k, out)
	if err != ErrBufferTooSmall {
		t.Error("expected error; got", err)
	}
	if length != 0 {
		t.Error("expected zero length; got", length)
	}

	// Construct a buffer that has the correct capacity but a smaller length.
	out = make([]byte, len(x)-Overhead)
	smallOut := out[:2]
	if len(smallOut) != 2 || cap(smallOut) != len(x)-Overhead {
		t.Error("invalid construction for test")
	}
	length, err = Decrypt(x, k, smallOut)
	if err != nil {
		t.Error("unexpected error:", err)
	}
	if length != len(x)-Overhead {
		t.Error("unexpected length; got", length)
	}
	if !bytes.Equal(m, smallOut[:len(x)-Overhead]) {
		t.Error("decrypted plaintext does not match original")
	}

	// Generate an incorrect key.
	ik := make([]byte, 32)
	Scramble(ik)

	// Attempt decryption with the incorrect key.
	length, err = Decrypt(x, ik, dm)
	if length != 0 {
		t.Error("expected length = 0; got", length)
	}
	if err != ErrDecryptionFailed {
		t.Error("expected error with incorrect key; got", err)
	}

	// Modify the ciphertext somewhat.
	for i := range x {
		if i%32 == 0 {
			x[i] = 0xdb
		}
	}

	// Attempt decryption of the invalid ciphertext with the correct key.
	length, err = Decrypt(x, k, dm)
	if length != 0 {
		t.Error("expected length = 0; got", length)
	}
	if err != ErrDecryptionFailed {
		t.Error("expected error with modified ciphertext; got", err)
	}

	// Generate a key of an invalid length.
	ik = make([]byte, 16)
	Scramble(ik)

	// Attempt encryption with the invalid key.
	ix, err := Encrypt(m, ik)
	if err != ErrInvalidKeyLength {
		t.Error("expected error with invalid key; got", err)
	}
	if ix != nil {
		t.Error("expected nil ciphertext; got", dm)
	}

	// Attempt decryption with the invalid key.
	length, err = Decrypt(x, ik, dm)
	if length != 0 {
		t.Error("expected length = 0; got", length)
	}
	if err != ErrInvalidKeyLength {
		t.Error("expected error with invalid key; got", err)
	}
}
