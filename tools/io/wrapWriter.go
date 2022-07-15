package io

import (
	"io"
)

type wrapWriter struct {
	w     io.Writer
	cur   int    // 记录当前行已经写入数量
	wrap  int    // 每行限制长度
	write []byte // 每行结尾追加数据,一般都是换行符
}

func NewWrapWriter(w io.Writer, wrap int, write []byte) io.Writer {
	return &wrapWriter{w: w, wrap: wrap, write: write}
}

func (w *wrapWriter) writeLine(b []byte) (n int, err error) {
	for tn, lb := 0, len(b); n < lb && err == nil; {
		tn, err = w.w.Write(b[n:])
		n += tn // 确保b中字节全部写入成功
	}
	if err == nil {
		_, err = w.w.Write(w.write)
	}
	return
}

func (w *wrapWriter) Write(b []byte) (n int, err error) {
	var nn int
	if now := w.wrap - w.cur; len(b) > now {
		nn, err = w.writeLine(b[:now])
		if n += nn; err != nil {
			return
		}
		b, w.cur = b[now:], 0

		for len(b) > w.wrap {
			nn, err = w.writeLine(b[:w.wrap])
			if n += nn; err != nil {
				return
			}
			b = b[w.wrap:]
		}
	}

	nn, err = w.w.Write(b)
	if n += nn; err != nil {
		return
	}
	w.cur += len(b)
	return
}
