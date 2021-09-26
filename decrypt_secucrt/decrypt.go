package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var filePath, passPhrase string
	switch len(os.Args) {
	case 3:
		passPhrase = os.Args[2]
		fallthrough
	case 2:
		filePath = os.Args[1]
	default:
		fmt.Printf("usage: %s file [pass]\n", os.Args[0])
		return
	}

	const (
		preHost = "S:\"Hostname\"="
		preUser = "S:\"Username\"="
		prePass = "S:\"Password V2\"=02:"
	)
	aesKey := sha256.Sum256([]byte(passPhrase))

	err := filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() || filepath.Ext(f.Name()) != ".ini" {
			return nil
		}

		realPath, err := filepath.Rel(filePath, path)
		if err != nil {
			return err
		}

		fr, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fr.Close()

		var (
			cnt     = 0
			data    [3]string
			scanner = bufio.NewScanner(fr)
		)
		for scanner.Scan() && cnt < 3 {
			line := scanner.Text()
			if i := strings.Index(line, preHost); i >= 0 {
				data[0] = line[i+len(preHost):]
				cnt++
			} else if i = strings.Index(line, preUser); i >= 0 {
				data[1] = line[i+len(preUser):] // 为了去掉文件BOM头部
				cnt++
			} else if i = strings.Index(line, prePass); i >= 0 {
				data[2], err = SecureCRTCryptoV2(aesKey[:], line[i+len(prePass):])
				if err != nil {
					return err
				}
				cnt++
			}
		}
		if cnt == 3 {
			fmt.Printf("%s [%s],[%s],[%s]\n", realPath, data[0], data[1], data[2])
		}
		return scanner.Err()
	})
	if err != nil {
		panic(err)
	}
}

func SecureCRTCryptoV2(key []byte, Ciphertext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	CipherSrc, err := hex.DecodeString(Ciphertext)
	if err != nil {
		return "", err
	}

	blockMode := cipher.NewCBCDecrypter(block, make([]byte, block.BlockSize()))
	blockMode.CryptBlocks(CipherSrc, CipherSrc)

	length := int(binary.LittleEndian.Uint32(CipherSrc))
	if length+4+sha256.Size > len(CipherSrc) {
		return "", errors.New("invalid size")
	}
	var (
		passByte   = CipherSrc[4 : 4+length]
		passSha256 = CipherSrc[4+length : 4+length+sha256.Size]
	)

	ok := false
	for i, v := range sha256.Sum256(passByte) {
		if passSha256[i] != v {
			ok = true
			break
		}
	}
	if ok {
		return "", errors.New("invalid ciphertext")
	}
	return string(passByte), nil
}
