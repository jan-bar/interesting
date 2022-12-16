package z7

import (
	"fmt"
	"strconv"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

//goland:noinspection GoUnusedConst,GoSnakeCaseUsage,SpellCheckingInspection
const (
	E_OK          = 0
	E_FAIL        = 0x80004005
	E_ABORT       = 0x80004004
	E_NOINTERFACE = 0x80004002 // No such interface supported
	E_OUTOFMEMORY = 0x8007000E
)

type ProPid uint64

// CPP\7zip\PropID.h
//goland:noinspection GoUnusedConst
const (
	KPidNoProperty ProPid = iota
	KPidMainSubFile
	KPidHandlerItemIndex
	KPidPath
	KPidName
	KPidExtension
	KPidIsDir
	KPidSize
	KPidPackSize
	KPidAttrib
	KPidCTime
	KPidATime
	KPidMTime
	KPidSolid
	KPidCommented
	KPidEncrypted
	KPidSplitBefore
	KPidSplitAfter
	KPidDictionarySize
	KPidCRC
	KPidType
	KPidIsAnti
	KPidMethod
	KPidHostOS
	KPidFileSystem
	KPidUser
	KPidGroup
	KPidBlock
	KPidComment
	KPidPosition
	KPidPrefix
	KPidNumSubDirs
	KPidNumSubFiles
	KPidUnpackVer
	KPidVolume
	KPidIsVolume
	KPidOffset
	KPidLinks
	KPidNumBlocks
	KPidNumVolumes
	KPidTimeType
	KPidBit64
	KPidBigEndian
	KPidCpu
	KPidPhySize
	KPidHeadersSize
	KPidChecksum
	KPidCharActs
	KPidVa
	KPidId
	KPidShortName
	KPidCreatorApp
	KPidSectorSize
	KPidPosixAttrib
	KPidSymLink
	KPidError
	KPidTotalSize
	KPidFreeSpace
	KPidClusterSize
	KPidVolumeName
	KPidLocalName
	KPidProvider
	KPidNtSecure
	KPidIsAltStream
	KPidIsAux
	KPidIsDeleted
	KPidIsTree
	KPidSha1
	KPidSha256
	KPidErrorType
	KPidNumErrors
	KPidErrorFlags
	KPidWarningFlags
	KPidWarning
	KPidNumStreams
	KPidNumAltStreams
	KPidAltStreamsSize
	KPidVirtualSize
	KPidUnpackSize
	KPidTotalPhySize
	KPidVolumeIndex
	KPidSubType
	KPidShortComment
	KPidCodePage
	KPidIsNotArcType
	KPidPhySizeCantBeDetected
	KPidZerosTailIsAllowed
	KPidTailSize
	KPidEmbeddedStubSize
	KPidNtReparse
	KPidHardLink
	KPidINode
	KPidStreamId
	KPidReadOnly
	KPidOutName
	KPidCopyLink
	KPidArcFileName
	KPidIsHash
	KPidChangeTime
	KPidUserId
	KPidGroupId
	KPidDeviceMajor
	KPidDeviceMinor
	KPidNumDefined
	KPidUserDefined ProPid = 0x10000
)

// 用于判断 PropVarIAnt.Vt 类型,然后根据具体类型解析 PropVarIAnt.Data
// https://learn.microsoft.com/zh-cn/windows/win32/api/wtypes/ne-wtypes-varenum
//goland:noinspection GoUnusedConst,GoSnakeCaseUsage,SpellCheckingInspection
const (
	VT_EMPTY            = 0
	VT_NULL             = 1
	VT_I2               = 2
	VT_I4               = 3
	VT_R4               = 4
	VT_R8               = 5
	VT_CY               = 6
	VT_DATE             = 7
	VT_BSTR             = 8
	VT_DISPATCH         = 9
	VT_ERROR            = 10
	VT_BOOL             = 11
	VT_VARIANT          = 12
	VT_UNKNOWN          = 13
	VT_DECIMAL          = 14
	VT_I1               = 16
	VT_UI1              = 17
	VT_UI2              = 18
	VT_UI4              = 19
	VT_I8               = 20
	VT_UI8              = 21
	VT_INT              = 22
	VT_UINT             = 23
	VT_VOID             = 24
	VT_HRESULT          = 25
	VT_PTR              = 26
	VT_SAFEARRAY        = 27
	VT_CARRAY           = 28
	VT_USERDEFINED      = 29
	VT_LPSTR            = 30
	VT_LPWSTR           = 31
	VT_RECORD           = 36
	VT_INT_PTR          = 37
	VT_UINT_PTR         = 38
	VT_FILETIME         = 64
	VT_BLOB             = 65
	VT_STREAM           = 66
	VT_STORAGE          = 67
	VT_STREAMED_OBJECT  = 68
	VT_STORED_OBJECT    = 69
	VT_BLOB_OBJECT      = 70
	VT_CF               = 71
	VT_CLSID            = 72
	VT_VERSIONED_STREAM = 73
	VT_BSTR_BLOB        = 0xfff
	VT_VECTOR           = 0x1000
	VT_ARRAY            = 0x2000
	VT_BYREF            = 0x4000
	VT_RESERVED         = 0x8000
	VT_ILLEGAL          = 0xffff
	VT_ILLEGALMASKED    = 0xfff
	VT_TYPEMASK         = 0xfff
)

type IUnknown struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

// PropVarIAnt 定义在: PropIdlBase.h
// https://learn.microsoft.com/zh-cn/windows/win32/api/propidlbase/ns-propidlbase-propvariant
type PropVarIAnt struct {
	Vt         uint16
	wReserved1 uint16
	wReserved2 uint16
	wReserved3 uint16
	Data       uintptr // 对应联合体,根据Vt类型进行解析
}

func (p *PropVarIAnt) String() string {
	switch p.Vt {
	case VT_EMPTY:
		return ""
	case VT_BSTR: // 对应dll中双字节字符串
		return UintPtrWToString(p.Data)
	case VT_UI1, VT_UI2, VT_UI4, VT_UI8:
		return strconv.FormatUint(uint64(p.Data), 10)
	case VT_FILETIME:
		fileData := uint64(p.Data)
		fileTime := &syscall.Filetime{
			LowDateTime:  uint32(fileData & 0xffffffff),
			HighDateTime: uint32(fileData >> 32 & 0xffffffff),
		}
		return time.Unix(0, fileTime.Nanoseconds()).String()
	case VT_I1, VT_I2, VT_I4, VT_I8:
		return strconv.FormatInt(int64(p.Data), 10)
	case VT_BOOL:
		if uint16(p.Data) != 0 {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("?:%d", p.Vt)
	}
}

func UintPtrWToString(r uintptr) string {
	//goland:noinspection GoVetUnsafePointer
	p := (*uint16)(unsafe.Pointer(r))
	if p == nil {
		return ""
	}

	n, end, add := 0, unsafe.Pointer(p), unsafe.Sizeof(*p)
	for *(*uint16)(end) != 0 {
		end = unsafe.Add(end, add)
		n++ // 计算该字符串以'\0'结束的长度
	}
	return string(utf16.Decode(unsafe.Slice(p, n)))
}
