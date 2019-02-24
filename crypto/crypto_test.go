package crypto_test

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/awnumar/memguard/crypto"
)

func TestCopy(t *testing.T) {
	a, err := crypto.GetRandBytes(8)
	if err != nil {
		t.Error(err)
	}
	b, err := crypto.GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}
	c, err := crypto.GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	// dst > src
	crypto.Copy(b, a)
	if !bytes.Equal(b[:8], a) {
		t.Error("incorrect copying")
	}

	// dst < src
	crypto.Copy(b, c)
	if !bytes.Equal(b, c[:16]) {
		t.Error("incorrect copying")
	}

	// dst = src
	b2, err := crypto.GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}
	crypto.Copy(b, b2)
	if !bytes.Equal(b, b2) {
		t.Error("incorrect copying")
	}
}

func TestMove(t *testing.T) {
	a, err := crypto.GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}
	b, err := crypto.GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	crypto.Move(a, b)
	if !bytes.Equal(b, make([]byte, 32)) {
		t.Error("src buffer was not wiped")
	}
}

func TestCompare(t *testing.T) {
	a, err := crypto.GetRandBytes(8)
	if err != nil {
		t.Error(err)
	}
	b, err := crypto.GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}
	c := make([]byte, 16)
	copy(c, b)

	// not equal
	if crypto.Equal(a, b) {
		t.Error("expected not equal")
	}

	// equal
	if !crypto.Equal(b, c) {
		t.Error("expected equal")
	}
}

func TestGetRandBytes(t *testing.T) {
	b, err := crypto.GetRandBytes(32)
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
		if base64.StdEncoding.EncodeToString(crypto.Hash([]byte(k))) != v {
			t.Error("digest doesn't match known values")
		}
	}
}

func TestMemClr(t *testing.T) {
	b, err := crypto.GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	crypto.MemClr(b)
	for i := range b {
		if b[i] != 0 {
			t.Error("memclr unsuccessful")
		}
	}
}

func TestMemSet(t *testing.T) {
	b, err := crypto.GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	crypto.MemSet(b, 0xdb)
	for i := range b {
		if b[i] != 0xdb {
			t.Error("memset unsuccessful")
		}
	}
}

func TestMemScr(t *testing.T) {
	b := make([]byte, 32)

	if err := crypto.MemScr(b); err != nil {
		t.Error(err)
	}
	if bytes.Equal(b, make([]byte, 32)) {
		t.Error("memscr unsuccessful")
	}
}

func TestSealOpen(t *testing.T) {
	// Declare the plaintext and the key.
	m, err := crypto.GetRandBytes(64)
	if err != nil {
		t.Error(err)
	}
	k, err := crypto.GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	// Encrypt the message.
	x, err := crypto.Seal(m, k)
	if err != nil {
		t.Error("expected no errors; got", err)
	}

	// Decrypt the message.
	dm := make([]byte, len(x)-crypto.Overhead)
	length, err := crypto.Open(x, k, dm)
	if err != nil {
		t.Error("expected no errors; got", err)
	}
	if length != len(x)-crypto.Overhead {
		t.Error("unexpected plaintext length; got", length)
	}

	// Verify that the plaintexts match.
	if !bytes.Equal(m, dm) {
		t.Error("decrypted plaintext does not match original")
	}

	// Attempt decryption /w buffer that is too small to hold the output.
	out := make([]byte, len(x)-crypto.Overhead-1)
	length, err = crypto.Open(x, k, out)
	if err != crypto.ErrBufferTooSmall {
		t.Error("expected error; got", err)
	}
	if length != 0 {
		t.Error("expected zero length; got", length)
	}

	// Construct a buffer that has the correct capacity but a smaller length.
	out = make([]byte, len(x)-crypto.Overhead)
	small_out := out[:2]
	if len(small_out) != 2 || cap(small_out) != len(x)-crypto.Overhead {
		t.Error("invalid construction for test")
	}
	length, err = crypto.Open(x, k, small_out)
	if err != nil {
		t.Error("unexpected error:", err)
	}
	if length != len(x)-crypto.Overhead {
		t.Error("unexpected length; got", length)
	}
	if !bytes.Equal(m, small_out[:len(x)-crypto.Overhead]) {
		t.Error("decrypted plaintext does not match original")
	}

	// Generate an incorrect key.
	ik, err := crypto.GetRandBytes(32)
	if err != nil {
		t.Error(err)
	}

	// Attempt decryption with the incorrect key.
	length, err = crypto.Open(x, ik, dm)
	if length != 0 {
		t.Error("expected length = 0; got", length)
	}
	if err != crypto.ErrDecryptionFailed {
		t.Error("expected error with incorrect key; got", err)
	}

	// Modify the ciphertext somewhat.
	for i := range x {
		if i%32 == 0 {
			x[i] = 0xdb
		}
	}

	// Attempt decryption of the invalid ciphertext with the correct key.
	length, err = crypto.Open(x, k, dm)
	if length != 0 {
		t.Error("expected length = 0; got", length)
	}
	if err != crypto.ErrDecryptionFailed {
		t.Error("expected error with modified ciphertext; got", err)
	}

	// Generate a key of an invalid length.
	ik, err = crypto.GetRandBytes(16)
	if err != nil {
		t.Error(err)
	}

	// Attempt encryption with the invalid key.
	ix, err := crypto.Seal(m, ik)
	if err != crypto.ErrInvalidKeyLength {
		t.Error("expected error with invalid key; got", err)
	}
	if ix != nil {
		t.Error("expected nil ciphertext; got", dm)
	}

	// Attempt decryption with the invalid key.
	length, err = crypto.Open(x, ik, dm)
	if length != 0 {
		t.Error("expected length = 0; got", length)
	}
	if err != crypto.ErrInvalidKeyLength {
		t.Error("expected error with invalid key; got", err)
	}
}
