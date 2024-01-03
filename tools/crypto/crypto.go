package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"hash"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

var salted = []byte("Salted__")

type AesMode uint8

const (
	Aes128 AesMode = iota
	Aes192
	Aes256
)

func genKeyIv(pass []byte, salt *[]byte, iter int, mode AesMode, md func() hash.Hash) (key, iv []byte, err error) {
	if *salt == nil {
		st := make([]byte, 8)
		if _, err = io.ReadFull(rand.Reader, st); err != nil {
			return
		}
		*salt = st
	}

	var keyLen int
	switch mode {
	case Aes128:
		keyLen = 16
	case Aes192:
		keyLen = 24
	case Aes256:
		keyLen = 32
	default:
		err = errors.New("mode not support")
		return
	}

	// make sure the generated key and iv are of sufficient length
	tmp := pbkdf2.Key(pass, *salt, iter, keyLen+aes.BlockSize, md)
	// get key and iv from result
	key, iv = tmp[:keyLen], tmp[keyLen:keyLen+aes.BlockSize]
	return
}

func AesCbcEnc(src io.Reader, dst io.Writer, pass []byte, iter int, mode AesMode, md func() hash.Hash) error {
	var salt []byte
	key, iv, err := genKeyIv(pass, &salt, iter, mode, md)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	bs := block.BlockSize()
	enc := cipher.NewCBCEncrypter(block, iv)

	var buf bytes.Buffer
	n, err := buf.ReadFrom(src)
	if err != nil {
		return err
	}

	for i, pl := 0, bs-int(n)%bs; i < pl; i++ {
		buf.WriteByte(byte(pl))
	}

	tmp := make([]byte, buf.Len())
	enc.CryptBlocks(tmp, buf.Bytes())

	buf.Reset()
	buf.Write(salted)
	buf.Write(salt)
	buf.Write(tmp)

	_, err = dst.Write(buf.Bytes())
	return err
}

func AesCbcDec(src io.Reader, dst io.Writer, pass []byte, iter int, mode AesMode, md func() hash.Hash) error {
	salt := append([]byte{}, salted...)
	for bytes.Equal(salt, salted) {
		_, err := io.ReadFull(src, salt)
		if err != nil {
			return err
		}
	}

	key, iv, err := genKeyIv(pass, &salt, iter, mode, md)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	bs := block.BlockSize()
	dec := cipher.NewCBCDecrypter(block, iv)

	org, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	ls := len(org)
	if ls == 0 || ls%bs != 0 {
		return errors.New("crypto/cipher: input error")
	}

	tmp := make([]byte, ls)
	dec.CryptBlocks(tmp, org)

	if pd := ls - int(tmp[ls-1]); pd > 0 {
		tmp = tmp[:pd]
	}

	_, err = dst.Write(tmp)
	return err
}
