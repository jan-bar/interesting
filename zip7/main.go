package main

import (
	"flag"
	"fmt"
	"os"
	"unsafe"

	"interesting/zip7/z7"
)

/*
本项目查询了多方资料，最终只做到读取非加密的7z文件,加密的回调一直没调通

个人感觉调用dll方式操作7z文件简直不太行，而且网上也没多少调用7z.dll的资料
由于是C++导出的dll，不想C导出dll那样简单，但是7z也有C语言版本，可以编译dll和so

有关压缩还是用下面那些纯Go实现的吧，或者用标准库的zip也够用了，这份研究暂时封存吧
https://github.com/bodgit/sevenzip
https://github.com/ulikunitz/xz

go run main.go -7z d:\7z.dll -f d:\xxx.7z
*/

func main() {
	z7dll := flag.String("7z", "7z.dll", "")
	z7zip := flag.String("f", "xxx.7z", "")
	flag.Parse()

	z7.LoadDll(*z7dll)

	var iInArchive *z7.IInArchive
	err := z7.CreateObject(z7.CLSIDFormat7z.ToGuid(), z7.IIDIInArchive.ToGuid(),
		uintptr(unsafe.Pointer(&iInArchive)))
	if err != nil {
		panic(err)
	}

	// 目前制作到打开不加密的压缩包
	fr, err := os.Open(*z7zip)
	if err != nil {
		panic(err)
	}
	defer fr.Close()

	inStream := z7.NewIInStream(fr)

	callback := z7.NewIArchiveOpenCallback("123")

	var maxCheckStartPosition uint64 = 1 << 23
	err = iInArchive.Open(inStream, &maxCheckStartPosition, callback)
	if err != nil {
		panic(err)
	}
	defer iInArchive.Close()

	num, err := iInArchive.GetNumberOfItems()
	if err != nil {
		panic(err)
	}
	fmt.Println(num)

	p := new(z7.PropVarIAnt)
	for i := uint32(0); i < num; i++ {
		err = iInArchive.GetProperty(i, z7.KPidPath, p)
		if err != nil {
			panic(err)
		}
		fmt.Println(p, unsafe.Sizeof(p))

		err = iInArchive.GetProperty(i, z7.KPidCRC, p)
		if err != nil {
			panic(err)
		}
		fmt.Println(p, unsafe.Sizeof(p))

		err = iInArchive.GetProperty(i, z7.KPidMTime, p)
		if err != nil {
			panic(err)
		}
		fmt.Println(p, unsafe.Sizeof(p))

		err = iInArchive.GetProperty(i, z7.KPidSize, p)
		if err != nil {
			panic(err)
		}
		fmt.Println(p, unsafe.Sizeof(p))

		err = iInArchive.GetProperty(i, z7.KPidPackSize, p)
		if err != nil {
			panic(err)
		}
		fmt.Println(p, unsafe.Sizeof(p))
	}
}
