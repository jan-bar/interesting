/*Package LzmaSpec

本文件依赖LzmaUtil.c,是吧lzma2201\C\Util\Lzma\LzmaUtil.c里面改出来的
这个是Windows和Linux下都可以编译成功,在include目录下把lzma2201\C这里面的对应.h拷贝的
执行如下就可以编译出libLzmaUtil.a
mingw32-make.exe -f makefile.util SOURCE=D:\code_project\c\lzma2201\C
*/
package LzmaSpec

/*
#cgo LDFLAGS: -L. -lLzmaUtil
#include <stdlib.h>
#include "include/LzmaUtil.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

type UtilType byte

const (
	UtilDec UtilType = 1
	UtilEnc UtilType = 2
)

func LzmaCompressUtil(src, dst string, mode UtilType) error {
	_, err := os.Stat(src)
	if err != nil {
		return err
	}
	dir := filepath.Dir(dst)
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.New(dir + " is not dir")
	}
	srcC, dstC := C.CString(src), C.CString(dst)
	defer C.free(unsafe.Pointer(srcC))
	defer C.free(unsafe.Pointer(dstC))
	ret := int(C.lzmaCompressUtil(srcC, dstC, C.int(mode)))
	if ret != SzOk {
		return fmt.Errorf("LzmaCompressUtil ret: %d", ret)
	}
	return nil
}
