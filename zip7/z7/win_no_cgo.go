//go:build windows && !cgo

package z7

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"
)

//goland:noinspection GoUnusedGlobalVariable,SpellCheckingInspection
var (
	z7dll = syscall.NewLazyDLL(`D:\Download\7z2201-extra\x64\7za.dll`)

	// createDecoder       = z7dll.NewProc("CreateDecoder")
	// createEncoder       = z7dll.NewProc("CreateEncoder")
	createObject = z7dll.NewProc("CreateObject")
	// getHandlerProperty  = z7dll.NewProc("GetHandlerProperty")
	// getHandlerProperty2 = z7dll.NewProc("GetHandlerProperty2")
	// getHashers          = z7dll.NewProc("GetHashers")
	// getIsArc            = z7dll.NewProc("GetIsArc")
	// getMethodProperty   = z7dll.NewProc("GetMethodProperty")
	getNumberOfFormats = z7dll.NewProc("GetNumberOfFormats")
	getNumberOfMethods = z7dll.NewProc("GetNumberOfMethods")
	// setCaseSensitive    = z7dll.NewProc("SetCaseSensitive")
	// setCodecs           = z7dll.NewProc("SetCodecs")
	// setLargePageMode    = z7dll.NewProc("SetLargePageMode")
)

type GUID string

const (
	IIDIUnknown GUID = "00000000-0000-0000-C000000000000046"

	CLSIDFormatZip   GUID = "23170F69-40C1-278A-1000000110010000"
	CLSIDFormatBzip2 GUID = "23170F69-40C1-278A-1000000110020000"
	CLSIDFormat7z    GUID = "23170F69-40C1-278A-1000000110070000"
	CLSIDFormatXz    GUID = "23170F69-40C1-278A-10000001100C0000"
	CLSIDFormatTar   GUID = "23170F69-40C1-278A-1000000110EE0000"
	CLSIDFormatGzip  GUID = "23170F69-40C1-278A-1000000110EF0000"

	IIDIInArchive  GUID = "23170F69-40C1-278A-0000000600600000"
	IIDIOutArchive GUID = "23170F69-40C1-278A-0000000600A00000"

	ICryptoGetTextPassword GUID = "23170F69-40C1-278A-0000000500100000"
)

var hexTable = map[byte]byte{'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5,
	'6': 6, '7': 7, '8': 8, '9': 9, 'A': 10, 'a': 10, 'B': 11, 'b': 11,
	'C': 12, 'c': 12, 'D': 13, 'd': 13, 'E': 14, 'e': 14, 'F': 15, 'f': 15}

func (g GUID) ToGuid() *syscall.GUID {
	const guidLen = 16 // GUID 内存占用为16

	//goland:noinspection GoRedundantConversion
	var (
		guid syscall.GUID // buf复用guid内存,num确保赋值大小端正确
		buf  = unsafe.Slice((*byte)(unsafe.Pointer(&guid)), guidLen)
		num  = []int{3, 2, 1, 0, 5, 4, 7, 6, 8, 9, 10, 11, 12, 13, 14, 15}
	)
	// 下面方式读取,只要可以读到连续16个hex字符就行,不限制字符串格式
	for i, j := 0, 0; i < len(g) && j < guidLen; {
		h, ok := hexTable[g[i]]
		i++
		if ok {
			l, ok := hexTable[g[i]]
			i++
			if ok {
				buf[num[j]] = h<<4 | l
				j++
			}
		}
	}
	return &guid
}

func GuidToGuid(guid *syscall.GUID) GUID {
	gs := fmt.Sprintf("%08X-%04X-%04X-%02X%02X%02X%02X%02X%02X%02X%02X",
		guid.Data1, guid.Data2, guid.Data3,
		guid.Data4[0], guid.Data4[1], guid.Data4[2], guid.Data4[3],
		guid.Data4[4], guid.Data4[5], guid.Data4[6], guid.Data4[7])
	return GUID(gs)
}

type (
	IInStreamVtable struct {
		IUnknown
		Read uintptr
		Seek uintptr
	}
	IInStream struct {
		Vtable  *IInStreamVtable
		NumRefs int // 自定义数据必须在Vtable之后
		Rs      io.ReadSeeker
	}
)

func NewIInStream(rs io.ReadSeeker) *IInStream {
	return &IInStream{
		Vtable: &IInStreamVtable{
			IUnknown: IUnknown{ // 这个逻辑是抄的,具体作用需要懂C++的大佬
				QueryInterface: syscall.NewCallback(func(self, iid, outObj uintptr) uintptr {
					return E_NOINTERFACE
				}),
				AddRef: syscall.NewCallback(func(self uintptr) uintptr {
					//goland:noinspection GoVetUnsafePointer
					z := (*IInStream)(unsafe.Pointer(self))
					z.NumRefs++
					return uintptr(z.NumRefs)
				}),
				Release: syscall.NewCallback(func(self uintptr) uintptr {
					//goland:noinspection GoVetUnsafePointer
					z := (*IInStream)(unsafe.Pointer(self))
					z.NumRefs--
					return uintptr(z.NumRefs)
				}),
			},
			Read: syscall.NewCallback(func(self, data, size, processedSize uintptr) uintptr {
				//goland:noinspection GoVetUnsafePointer,GoRedundantConversion
				var (
					fs = (*IInStream)(unsafe.Pointer(self))
					dt = unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))
					ps = (*uint32)(unsafe.Pointer(processedSize))
				)
				n, err := fs.Rs.Read(dt)
				if ps != nil {
					*ps = uint32(n)
				}
				if err != nil {
					return E_FAIL // todo 思考如何返回准确错误
				}
				return E_OK
			}),
			Seek: syscall.NewCallback(func(self, offset, seekOrigin, newPosition uintptr) uintptr {
				//goland:noinspection GoVetUnsafePointer
				var (
					fs = (*IInStream)(unsafe.Pointer(self))
					ps = (*uint64)(unsafe.Pointer(newPosition))
				)
				n, err := fs.Rs.Seek(int64(offset), int(seekOrigin))
				if ps != nil {
					*ps = uint64(n)
				}
				if err != nil {
					return E_FAIL // todo 思考如何返回准确错误
				}
				return E_OK
			}),
		},
		NumRefs: 1,
		Rs:      rs,
	}
}

type (
	IArchiveOpenCallbackVtable struct {
		IUnknown
		SetTotal     uintptr
		SetCompleted uintptr // 这两个官方例子啥也没处理,直接返回0
	}
	IArchiveOpenCallback struct {
		Vtable   *IArchiveOpenCallbackVtable
		NumRefs  int
		Password *string
	}
)

type IUnknownI struct {
	RawVTable *interface{}
}

type (
	ICryptoGetTextPasswordIV struct {
		IUnknown
		CryptoGetTextPassword uintptr
	}

	ICryptoGetTextPasswordI struct {
		Vtable *ICryptoGetTextPasswordIV
	}
)

func init() {
	// Library
	liboleaut32 := syscall.NewLazyDLL("oleaut32.dll")

	// Functions
	sysAllocString = liboleaut32.NewProc("SysAllocString")
	// sysFreeString = liboleaut32.NewProc("SysFreeString")
	// sysStringLen = liboleaut32.NewProc("SysStringLen")
}

var sysAllocString *syscall.LazyProc

func SysAllocString(s string) *int16 {
	ret, _, _ := syscall.Syscall(sysAllocString.Addr(), 1,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s))),
		0,
		0)

	return (*int16)(unsafe.Pointer(ret))
}

func NewIArchiveOpenCallback(pass ...string) *IArchiveOpenCallback {
	ioc := &IArchiveOpenCallback{
		Vtable: &IArchiveOpenCallbackVtable{
			IUnknown: IUnknown{
				QueryInterface: syscall.NewCallback(func(self *IUnknownI, iid uintptr, outObj **IUnknownI) uintptr {
					return E_NOINTERFACE // 可以打开不加密的压缩包

					//goland:noinspection GoVetUnsafePointer
					var (
						z  = (*IArchiveOpenCallback)(unsafe.Pointer(self))
						id = (*syscall.GUID)(unsafe.Pointer(iid))
					)
					*outObj = nil
					if GuidToGuid(id) == ICryptoGetTextPassword {
						xx := &ICryptoGetTextPasswordI{
							Vtable: &ICryptoGetTextPasswordIV{
								CryptoGetTextPassword: syscall.NewCallback(func(self, pass uintptr) uintptr {
									xxx := SysAllocString("123")

									a := (*int16)(unsafe.Pointer(pass))
									a = xxx

									fmt.Println(*a)
									return E_OK
								}),
							},
						}
						*outObj = (*IUnknownI)(unsafe.Pointer(xx))
					} else {
						return E_NOINTERFACE
					}
					z.NumRefs++
					return E_OK
				}),
				AddRef: syscall.NewCallback(func(self uintptr) uintptr {
					//goland:noinspection GoVetUnsafePointer
					z := (*IArchiveOpenCallback)(unsafe.Pointer(self))
					z.NumRefs++
					return uintptr(z.NumRefs)
				}),
				Release: syscall.NewCallback(func(self uintptr) uintptr {
					//goland:noinspection GoVetUnsafePointer
					z := (*IArchiveOpenCallback)(unsafe.Pointer(self))
					z.NumRefs--
					return uintptr(z.NumRefs)
				}),
			},
			SetTotal: syscall.NewCallback(func(self, files, bytes uintptr) uintptr {
				// SetTotal(const UInt64 *files, const UInt64 *bytes);
				return E_OK
			}),
			SetCompleted: syscall.NewCallback(func(self, files, bytes uintptr) uintptr {
				// SetCompleted(const UInt64 *files, const UInt64 *bytes)
				return E_OK
			}),
		},
		NumRefs: 1,
	}
	if len(pass) > 0 {
		ioc.Password = &pass[0] // 当传入密码时,需要赋值
	}
	return ioc
}

type (
	IInArchiveVtable struct {
		IUnknown
		Open                         uintptr // dll调用打开逻辑
		Close                        uintptr // dll调用关闭逻辑
		GetNumberOfItems             uintptr // 获取压缩文件总数
		GetProperty                  uintptr // 获取指定索引文件的详细属性
		Extract                      uintptr // 解压逻辑
		GetArchiveProperty           uintptr
		GetNumberOfProperties        uintptr
		GetPropertyInfo              uintptr
		GetNumberOfArchiveProperties uintptr
		GetArchivePropertyInfo       uintptr
	}
	IInArchive struct {
		Vtable *IInArchiveVtable
	}
)

func (ia *IInArchive) Open(stream *IInStream, maxCheckStartPosition *uint64,
	openArchiveCallback *IArchiveOpenCallback) error {
	ret, _, _ := syscall.SyscallN(ia.Vtable.Open,
		uintptr(unsafe.Pointer(ia)),
		uintptr(unsafe.Pointer(stream)),
		uintptr(unsafe.Pointer(maxCheckStartPosition)),
		uintptr(unsafe.Pointer(openArchiveCallback)),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

func (ia *IInArchive) Close() error {
	ret, _, _ := syscall.SyscallN(ia.Vtable.Close, uintptr(unsafe.Pointer(ia)))
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

func (ia *IInArchive) GetNumberOfItems() (numItems uint32, err error) {
	ret, _, _ := syscall.SyscallN(ia.Vtable.GetNumberOfItems,
		uintptr(unsafe.Pointer(ia)), uintptr(unsafe.Pointer(&numItems)))
	if ret != 0 {
		err = syscall.Errno(ret)
	}
	return
}

func (ia *IInArchive) GetProperty(index uint32,
	proPid ProPid, prop *PropVarIAnt) error {
	ret, _, _ := syscall.SyscallN(ia.Vtable.GetProperty,
		uintptr(unsafe.Pointer(ia)),
		uintptr(index), uintptr(proPid),
		uintptr(unsafe.Pointer(prop)),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

func CreateObject(cls, iid *syscall.GUID, outObject uintptr) error {
	ret, _, _ := createObject.Call(
		uintptr(unsafe.Pointer(cls)),
		uintptr(unsafe.Pointer(iid)), outObject)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

func GetNumberOfFormats() (num uint32, err error) {
	ret, _, _ := getNumberOfFormats.Call(uintptr(unsafe.Pointer(&num)))
	if ret != 0 {
		err = syscall.Errno(ret)
	}
	return
}

func GetNumberOfMethods() (num uint32, err error) {
	ret, _, _ := getNumberOfMethods.Call(uintptr(unsafe.Pointer(&num)))
	if ret != 0 {
		err = syscall.Errno(ret)
	}
	return
}
