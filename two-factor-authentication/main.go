package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	add := flag.String("a", "", `-a "name secretKey"`)
	size := flag.Int("s", 6, "generate [6,7,8]-digit code")
	list := flag.Bool("l", false, "list keys")
	flag.Parse()

	data := make(map[string]string)
	err := rwKeyChain(data, true)
	if err != nil {
		log.Fatal(err)
	}

	if *list {
		var va []string
		for k := range data {
			va = append(va, k)
		}
		sort.Strings(va)
		fmt.Println(strings.Join(va, "\n"))
		return
	}

	if as := strings.Fields(*add); len(as) == 2 {
		_, err = decBase32(&as[1])
		if err != nil {
			log.Fatal(err)
		}
		data[strings.TrimSpace(as[0])] = as[1]

		err = rwKeyChain(data, false)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// 30秒内时间戳,30秒内生成的code相同
	tn := uint64(time.Now().Unix() / 30)
	for _, name := range flag.Args() {
		key, ok := data[name]
		if !ok {
			continue
		}

		code, err := GenTOTP(key, tn, *size)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s %d\n", name, code)
	}
}

func rwKeyChain(data map[string]string, r bool) error {
	path, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path = filepath.Join(path, ".2fa")

	if r {
		fr, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		err = json.NewDecoder(fr).Decode(&data)
		_ = fr.Close()
		return err
	}

	fw, err := os.Create(path)
	if err != nil {
		return err
	}
	err = json.NewEncoder(fw).Encode(data)
	_ = fw.Close()
	return err
}

func decBase32(s *string) ([]byte, error) {
	*s = strings.ToUpper(strings.TrimSpace(*s))
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(*s)
}

func GenTOTP(secretKey string, counter uint64, digit int) (uint32, error) {
	secretBytes, err := decBase32(&secretKey)
	if err != nil {
		return 0, err
	}

	// bytes. Then a 20-byte SHA-1 hash is calculated from the byte slice
	hash := hmac.New(sha1.New, secretBytes)
	err = binary.Write(hash, binary.BigEndian, counter)
	if err != nil {
		return 0, err
	}
	h := hash.Sum(nil) // Calculate 20-byte SHA-1 digest

	// AND the SHA-1 with 0x0F (15) to get a single-digit offset
	offset := h[len(h)-1] & 0x0F

	// Truncate the SHA-1 by the offset and convert it into a 32-bit
	// unsigned int. AND the 32-bit int with 0x7FFFFFFF (2147483647)
	// to get a 31-bit unsigned int.
	truncatedHash := binary.BigEndian.Uint32(h[offset:]) & 0x7FFFFFFF

	var mod uint32 = 1_000_000 // [6,10] len digit
	for i := 6; i < 10 && i < digit; i++ {
		mod *= 10
	}
	return truncatedHash % mod, nil
}
