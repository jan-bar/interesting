package LzmaSpec

import (
	"fmt"
	"io"
)

// from 7zTypes.h
//goland:noinspection GoUnusedConst
const (
	SzOk               = 0
	SzErrorData        = 1
	SzErrorMem         = 2
	SzErrorCrc         = 3
	SzErrorUnsupported = 4
	SzErrorParam       = 5
	SzErrorInputEOF    = 6
	SzErrorOutputEOF   = 7
	SzErrorRead        = 8
	SzErrorWrite       = 9
	SzErrorProgress    = 10
	SzErrorFail        = 11
	SzErrorThread      = 12
	SzErrorArchive     = 16
	SzErrorNoArchive   = 17

	lzmaPropsSize = uint64(5)
	sizeHeader    = 8
)

func getLenBytes(byt []byte) uint64 {
	var size uint64
	for i := 0; i < sizeHeader; i++ {
		size |= uint64(byt[i]) << (8 * i)
	}
	return size
}

func setLenBytes(byt []byte, size uint64) {
	for i := 0; i < sizeHeader; i++ {
		byt[i] = byte(size >> (8 * i))
	}
}

type UseType byte

const (
	UseDll UseType = 1
	UseCgo UseType = 2
)

// LzmaCompress write {[5]byte porps} + {[8]byte fileLen} + {data}
func LzmaCompress(r io.Reader, w io.Writer, useType UseType) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	var (
		srcLen    = uint64(len(src))
		dstLen    = srcLen + srcLen/3 + 128 // 网上找了很多地方,都是这样算出来的
		dst       = make([]byte, lzmaPropsSize+sizeHeader+dstLen)
		propsSize = lzmaPropsSize
		ret       = -1
	)
	setLenBytes(dst[lzmaPropsSize:], srcLen) // 保存源文件长度

	switch useType { // outProps也需要保存
	case UseDll:
		err = lzmaCompressDll(dst[lzmaPropsSize+sizeHeader:], &dstLen, src, srcLen,
			dst[:lzmaPropsSize], &propsSize, 9, 1<<24, 3, 0, 2, 32, 2)
		if err != nil {
			return err
		}
	case UseCgo:
		ret = lzmaCompressCgo(dst[lzmaPropsSize+sizeHeader:], &dstLen, src, srcLen,
			dst[:lzmaPropsSize], &propsSize, 9, 1<<24, 3, 0, 2, 32, 2)
	default:
		return fmt.Errorf("useType error: %d", useType)
	}
	if ret != SzOk {
		return fmt.Errorf("lzmaCompress ret: %d", ret)
	}
	if propsSize != lzmaPropsSize {
		return fmt.Errorf("propsSize error")
	}
	dstLen += lzmaPropsSize + sizeHeader
	ret, err = w.Write(dst[:dstLen])
	if uint64(ret) != dstLen && err == nil {
		return io.ErrShortWrite
	}
	return err
}

// LzmaUnCompress read {[5]byte porps} + {[8]byte fileLen} + {data}
func LzmaUnCompress(r io.Reader, w io.Writer, useType UseType) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	var (
		dstLen = getLenBytes(src[lzmaPropsSize:]) // 读取源文件大小
		dst    = make([]byte, dstLen)             // 申请资源
		ret    = -1
		srcLen = uint64(len(src)) - lzmaPropsSize - sizeHeader // 去掉头部
	)
	switch useType { // 使用r中读到的props传递参数
	case UseDll:
		err = lzmaUnCompressDll(dst, &dstLen, src[lzmaPropsSize+sizeHeader:],
			&srcLen, src[:lzmaPropsSize], lzmaPropsSize)
		if err != nil {
			return err
		}
	case UseCgo:
		ret = lzmaUnCompressCgo(dst, &dstLen, src[lzmaPropsSize+sizeHeader:],
			&srcLen, src[:lzmaPropsSize], lzmaPropsSize)
	default:
		return fmt.Errorf("useType error: %d", useType)
	}
	if ret != SzOk {
		return fmt.Errorf("lzmaUnCompress ret: %d", ret)
	}
	ret, err = w.Write(dst[:dstLen])
	if uint64(ret) != dstLen && err == nil {
		return io.ErrShortWrite
	}
	return err
}
