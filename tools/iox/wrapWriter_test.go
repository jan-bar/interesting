package iox

import (
	rand2 "crypto/rand"
	"encoding/base64"
	"math/rand"
	"os"
	"testing"
	"time"
)

// go test -v -run TestWrapWriter
func TestWrapWriter(t *testing.T) {
	lw := base64.NewEncoder(base64.StdEncoding,
		NewWrapWriter(os.Stdout, 76, []byte{'\n'}))

	buf := make([]byte, 512)
	rand.Seed(time.Now().UnixNano())
	for {
		n := rand.Intn(80)
		_, _ = rand2.Read(buf[:n])

		// 随机写入长度数据,确保输出在wrap长度时追加write数据即可
		// 一般用来追加换行符
		_, err := lw.Write(buf[:n])
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)
	}
}
