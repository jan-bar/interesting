package main

import (
	"fmt"
	"os"
	"unsafe"

	"interesting/zip7/z7"
)

func main() {
	var iInArchive *z7.IInArchive
	err := z7.CreateObject(z7.CLSIDFormat7z.ToGuid(), z7.IIDIInArchive.ToGuid(),
		uintptr(unsafe.Pointer(&iInArchive)))
	if err != nil {
		panic(err)
	}

	fr, err := os.Open(`D:\code_project\python\project\pylib7zip\tests\complex\complex.7z`)
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
