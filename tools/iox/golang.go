package iox

import (
	"io"
)

// copy encoding/pem/pem.go#lineBreaker
// 相比我之前的 wrapWriter , 官方 lineBreaker 使用了line缓存
// 也就是每次调用底层Write时,一定是一行的数据,而我的 wrapWriter 则是有多少算多少
// 当遇到需要换行时,会调用多次底层Write
type lineBreaker struct {
	line   []byte    // 缓存一行的数据
	used   int       // 记录当前行已使用个数
	length int       // 限定长度
	out    io.Writer // 写入流
	nl     []byte    // 每行末尾追加数据,一般都是换行
}

func NewLineBreaker(w io.Writer, length int, nl []byte) io.Writer {
	return &lineBreaker{
		line:   make([]byte, length),
		length: length,
		out:    w,
		nl:     nl,
	}
}

func (l *lineBreaker) Write(b []byte) (n int, err error) {
	if l.used+len(b) < l.length {
		copy(l.line[l.used:], b)
		l.used += len(b)
		return len(b), nil
	}

	n, err = l.out.Write(l.line[:l.used])
	if err != nil {
		return
	}
	excess := l.length - l.used
	l.used = 0

	n, err = l.out.Write(b[:excess])
	if err != nil {
		return
	}

	n, err = l.out.Write(l.nl)
	if err != nil {
		return
	}

	return l.Write(b[excess:])
}

func (l *lineBreaker) Close() (err error) {
	if l.used > 0 {
		_, err = l.out.Write(l.line[:l.used])
		if err != nil {
			return
		}
		_, err = l.out.Write(l.nl)
	}
	return
}
