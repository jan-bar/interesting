/*Package LzmaSpec

注意本程序只适用于Windows下用,我已经试过Linux下编译不过
本程序和使用LZMA.dll方式的区别就是,使用静态编译,不需要依赖dll文件
执行如下命令,指定编译目录可以编译产生liblzma.a的静态库
mingw32-make.exe -f makefile.cgo SOURCE=D:\code_project\c\lzma2201\C
记得修改makefile里面的SOURCE路径,我已经将编译号的liblzma.a提交git了,大家随意使用
注意需要复制一系列头文件
*/
package LzmaSpec

/*
#cgo LDFLAGS: -L. -llzma
#include "include/LzmaLib.h"
*/
import "C"
import (
	"unsafe"
)

func lzmaCompressCgo(dst []byte, dstLen *uint64, src []byte, srcLen uint64,
	outProps []byte, outPropsSize *uint64,
	level, dictSize, lc, lp, pb, fb, numThreads uint32) int {
	return int(C.LzmaCompress(
		(*C.uchar)(unsafe.Pointer(&dst[0])), (*C.size_t)(unsafe.Pointer(dstLen)),
		(*C.uchar)(unsafe.Pointer(&src[0])), C.size_t(srcLen),
		(*C.uchar)(unsafe.Pointer(&outProps[0])),
		(*C.size_t)(unsafe.Pointer(outPropsSize)),
		C.int(level), C.unsigned(dictSize), C.int(lc), C.int(lp), C.int(pb),
		C.int(fb), C.int(numThreads)))
}

func lzmaUnCompressCgo(dst []byte, dstLen *uint64, src []byte, srcLen *uint64,
	props []byte, propsSize uint64) int {
	return int(C.LzmaUncompress(
		(*C.uchar)(unsafe.Pointer(&dst[0])), (*C.size_t)(unsafe.Pointer(dstLen)),
		(*C.uchar)(unsafe.Pointer(&src[0])), (*C.SizeT)(unsafe.Pointer(srcLen)),
		(*C.uchar)(unsafe.Pointer(&props[0])), C.size_t(propsSize)))
}
