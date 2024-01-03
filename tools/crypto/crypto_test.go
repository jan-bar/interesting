package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"os/exec"
	"testing"
)

func TestAesOpenssl(t *testing.T) {
	src := make([]byte, 1020) // test data
	_, err := io.ReadFull(rand.Reader, src)
	if err != nil {
		t.Fatal(err)
	}

	pb := make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, pb)
	if err != nil {
		t.Fatal(err)
	}

	// the openssl command line must have a password that can be printed
	pass := base64.StdEncoding.EncodeToString(pb)

	var buf, buf1 bytes.Buffer

	// code enc
	err = AesCbcEnc(bytes.NewReader(src), &buf, []byte(pass),
		400, Aes256, sha512.New)
	if err != nil {
		t.Fatal(err)
	}

	// command dec
	cmd := exec.Command("openssl", "aes-256-cbc", "-k", pass,
		"-d", "-md", "sha512", "-iter", "400", "-pbkdf2")
	cmd.Stdin = &buf
	cmd.Stdout = &buf1
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(src, buf1.Bytes()) {
		t.Fatal("not eq")
	}

	// command enc
	cmd = exec.Command("openssl", "aes-128-cbc", "-k", pass,
		"-e", "-md", "sha256", "-iter", "810", "-pbkdf2")
	cmd.Stdin = bytes.NewReader(src)
	buf.Reset()
	cmd.Stdout = &buf
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}

	buf1.Reset()

	// code dec
	err = AesCbcDec(&buf, &buf1, []byte(pass), 810, Aes128, sha256.New)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(src, buf1.Bytes()) {
		t.Fatal("not eq")
	}
}
