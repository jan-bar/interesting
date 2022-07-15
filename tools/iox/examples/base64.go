package main

import (
	"encoding/base64"
	"flag"
	"io"
	"os"

	"interesting/tools/iox"
)

func main() {
	dec := flag.Bool("d", false, "decode data")
	ignore := flag.Bool("i", false, "when decoding, ignore non-alphabet characters")
	wrap := flag.Int("w", 76, "wrap encoded lines after COLS character (default 76).\n  Use 0 to disable line wrapping")
	srcFile := flag.String("f", "-", "input file,\"-\" is stdin")
	dstFile := flag.String("o", "-", "output file,\"-\" is stdout")
	flag.Parse()

	var br io.Reader
	if *srcFile != "-" {
		fr, err := os.Open(*srcFile)
		if err != nil {
			panic(err)
		}
		//goland:noinspection GoUnhandledErrorResult
		defer fr.Close()
		br = fr
	} else {
		br = os.Stdin // 其他情况都从标准输入读取
	}

	var bo io.Writer
	if *dstFile != "-" {
		fw, err := os.Create(*dstFile)
		if err != nil {
			panic(err)
		}
		//goland:noinspection GoUnhandledErrorResult
		defer fw.Close()
		bo = fw
	} else {
		bo = os.Stdout // 其他情况都输出到标准输出
	}

	if *dec {
		if err := decode(br, bo, *ignore); err != nil {
			panic(err)
		}
	} else {
		if err := encode(br, bo, *wrap); err != nil {
			panic(err)
		}
	}
}

type ignoreRead struct {
	r       io.Reader
	mustMap map[byte]struct{}
}

func newIgnoreRead(r io.Reader, ig []byte) io.Reader {
	igr := &ignoreRead{
		r:       r,
		mustMap: make(map[byte]struct{}, len(ig)),
	}
	for _, b := range ig {
		igr.mustMap[b] = struct{}{}
	}
	return igr
}

func (i *ignoreRead) Read(b []byte) (n int, err error) {
	n, err = i.r.Read(b)
	if err != nil {
		return
	}
	b = b[:n]

	n = 0
	for _, v := range b {
		if _, ok := i.mustMap[v]; ok {
			b[n] = v // 只保留必须的字符
			n++
		}
	}
	return
}

func decode(r io.Reader, w io.Writer, ignore bool) error {
	if ignore {
		//goland:noinspection SpellCheckingInspection,忽略非base64以外的字符,拼接(base64.encodeStd + base64.StdPadding)
		igb := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=")
		r = newIgnoreRead(r, igb)
	}
	_, err := io.Copy(w, base64.NewDecoder(base64.StdEncoding, r))
	return err
}

func encode(r io.Reader, w io.Writer, wrap int) (err error) {
	if wrap > 0 {
		newLine := []byte{'\n'}
		w = iox.NewWrapWriter(w, wrap, newLine)

		defer func() {
			if err == nil {
				_, err = w.Write(newLine)
			}
		}()
	}

	bw := base64.NewEncoder(base64.StdEncoding, w)

	if _, err = io.Copy(bw, r); err != nil {
		return err
	}

	return bw.Close()
}
