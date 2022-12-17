/*Package LzmaSpec

[https://www.7-zip.org/a/lzma2201.7z]解压后lzma2201\C\Util\LzmaLib
使用vs2010打开LzmaLib.dsw,在配置管理器里面新建一个x64的release版本(产生dll比较小)
然后生成解决方案,会在C:\Util产生LZMA.dll文件

控制台命令生成lzma.dll
cmd /k "D:\xxx\vcvars64.bat" 在cmd中生效vs各项配置
按照 https://www.cnblogs.com/janbar/p/15644092.html 文章
cd lzma2201\C\Util\LzmaLib
nmake CPU=AMD64 NEW_COMPILER=1 MY_STATIC_LINK=1
就得到了lzma.dll

安装vs2010,可以用下面命令查看dll对外提供接口,只能查看函数名
dumpbin.exe -exports c:\Util\LZMA.dll
然后在这个\lzma2201\C\LzmaLib.h里面可以看到 LzmaCompress和 LzmaUncompress的定义
以及各种注意事项,下面用法就是从上面得到的
*/
package LzmaSpec

import (
	"syscall"
	"unsafe"
)

var (
	compress   *syscall.LazyProc
	unCompress *syscall.LazyProc
)

func LoadLzmaDll(file string) {
	dll := syscall.NewLazyDLL(file)
	compress = dll.NewProc("LzmaCompress")
	unCompress = dll.NewProc("LzmaUncompress")
}

/*
RAM requirements for LZMA:
  for compression:   (dictSize * 11.5 + 6 MB) + state_size
  for decompression: dictSize + state_size
    state_size = (4 + (1.5 << (lc + lp))) KB
    by default (lc=3, lp=0), state_size = 16 KB.

LZMA properties (5 bytes) format
    Offset Size  Description
      0     1    lc, lp and pb in encoded form.
      1     4    dictSize (little endian).

LzmaCompress
------------

outPropsSize -
     In:  the pointer to the size of outProps buffer; *outPropsSize = LZMA_PROPS_SIZE = 5.
     Out: the pointer to the size of written properties in outProps buffer; *outPropsSize = LZMA_PROPS_SIZE = 5.

  LZMA Encoder will use defult values for any parameter, if it is
  -1  for any from: level, loc, lp, pb, fb, numThreads
   0  for dictSize

level - compression level: 0 <= level <= 9;

  level dictSize algo  fb
    0:    16 KB   0    32
    1:    64 KB   0    32
    2:   256 KB   0    32
    3:     1 MB   0    32
    4:     4 MB   0    32
    5:    16 MB   1    32
    6:    32 MB   1    32
    7+:   64 MB   1    64

  The default value for "level" is 5.

  algo = 0 means fast method
  algo = 1 means normal method

dictSize - The dictionary size in bytes. The maximum value is
        128 MB = (1 << 27) bytes for 32-bit version
          1 GB = (1 << 30) bytes for 64-bit version
     The default value is 16 MB = (1 << 24) bytes.
     It's recommended to use the dictionary that is larger than 4 KB and
     that can be calculated as (1 << N) or (3 << N) sizes.

lc - The number of literal context bits (high bits of previous literal).
     It can be in the range from 0 to 8. The default value is 3.
     Sometimes lc=4 gives the gain for big files.

lp - The number of literal pos bits (low bits of current position for literals).
     It can be in the range from 0 to 4. The default value is 0.
     The lp switch is intended for periodical data when the period is equal to 2^lp.
     For example, for 32-bit (4 bytes) periodical data you can use lp=2. Often it's
     better to set lc=0, if you change lp switch.

pb - The number of pos bits (low bits of current position).
     It can be in the range from 0 to 4. The default value is 2.
     The pb switch is intended for periodical data when the period is equal 2^pb.

fb - Word size (the number of fast bytes).
     It can be in the range from 5 to 273. The default value is 32.
     Usually, a big number gives a little bit better compression ratio and
     slower compression process.

numThreads - The number of thereads. 1 or 2. The default value is 2.
     Fast mode (algo = 0) can use only 1 thread.

Out:
  destLen  - processed output size
Returns:
  SZ_OK               - OK
  SZ_ERROR_MEM        - Memory allocation error
  SZ_ERROR_PARAM      - Incorrect paramater
  SZ_ERROR_OUTPUT_EOF - output buffer overflow
  SZ_ERROR_THREAD     - errors in multithreading functions (only for Mt version)
*/
// MY_STDAPI LzmaCompress(unsigned char *dest, size_t *destLen, const unsigned char *src, size_t srcLen,
//   unsigned char *outProps, size_t *outPropsSize, /* *outPropsSize must be = 5 */
// int level,      /* 0 <= level <= 9, default = 5 */
// unsigned dictSize,  /* default = (1 << 24) */
// int lc,        /* 0 <= lc <= 8, default = 3  */
// int lp,        /* 0 <= lp <= 4, default = 0  */
// int pb,        /* 0 <= pb <= 4, default = 2  */
// int fb,        /* 5 <= fb <= 273, default = 32 */
// int numThreads /* 1 or 2, default = 2 */
// );
func lzmaCompressDll(dst []byte, dstLen *uint64, src []byte, srcLen uint64,
	outProps []byte, outPropsSize *uint64,
	level, dictSize, lc, lp, pb, fb, numThreads uint32) error {
	ret, _, _ := compress.Call(
		uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(dstLen)),
		uintptr(unsafe.Pointer(&src[0])), uintptr(srcLen),
		uintptr(unsafe.Pointer(&outProps[0])),
		uintptr(unsafe.Pointer(outPropsSize)),
		uintptr(level), uintptr(dictSize), uintptr(lc), uintptr(lp),
		uintptr(pb), uintptr(fb), uintptr(numThreads))
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

/*
LzmaUncompress
--------------
In:
  dest     - output data
  destLen  - output data size
  src      - input data
  srcLen   - input data size
Out:
  destLen  - processed output size
  srcLen   - processed input size
Returns:
  SZ_OK                - OK
  SZ_ERROR_DATA        - Data error
  SZ_ERROR_MEM         - Memory allocation arror
  SZ_ERROR_UNSUPPORTED - Unsupported properties
  SZ_ERROR_INPUT_EOF   - it needs more bytes in input buffer (src)

MY_STDAPI LzmaUncompress(unsigned char *dest, size_t *destLen, const unsigned char *src, SizeT *srcLen,
  const unsigned char *props, size_t propsSize);
*/
func lzmaUnCompressDll(dst []byte, dstLen *uint64, src []byte, srcLen *uint64,
	props []byte, propsSize uint64) error {
	ret, _, _ := unCompress.Call(
		uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(dstLen)),
		uintptr(unsafe.Pointer(&src[0])), uintptr(unsafe.Pointer(srcLen)),
		uintptr(unsafe.Pointer(&props[0])), uintptr(propsSize))
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}
