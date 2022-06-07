package tools

import (
	"regexp"
	"strconv"
	"strings"
)

var regUnicode = regexp.MustCompile(`\\u[\da-fA-F]{4}|\\U[\da-fA-F]{8}`)

func ReplaceUnicode(s string) string {
	return regUnicode.ReplaceAllStringFunc(s, parseUnicode)
}

// 找到unicode字符串,转换为utf8字节
func parseUnicode(s string) string {
	i, err := strconv.ParseUint(s[2:], 16, 32)
	if err != nil || i > 0x10ffff {
		return s // 转换错误,或者超出定义范围返回原字符串
	}

	/* https://en.wikipedia.org/wiki/UTF-8, 根据如下规则进行转换
	First code point  Last code point   Byte 1    Byte 2    Byte 3    Byte 4
	          U+0000           U+007F  0xxxxxxx
	          U+0080           U+07FF  110xxxxx  10xxxxxx
	          U+0800           U+FFFF  1110xxxx  10xxxxxx  10xxxxxx
	         U+10000   [nb 2]U+10FFFF  11110xxx  10xxxxxx  10xxxxxx  10xxxxxx
	*/
	const (
		mask0         = 0b01111111             // < 0x80    时的Byte 1掩码
		mask1, mask11 = 0b11000000, 0b00011111 // < 0x80    时的Byte 1掩码
		mask2, mask22 = 0b11100000, 0b00001111 // < 0x800   时的Byte 1掩码
		mask3, mask33 = 0b11110000, 0b00000111 // < 0x10000 时的Byte 1掩码
		mask4, mask44 = 0b10000000, 0b00111111 // 其他位的掩码
	)
	var str strings.Builder
	if i >= 0x10000 {
		str.WriteByte(mask3 | byte(i>>18&mask33))
		str.WriteByte(mask4 | byte(i>>12&mask44))
		str.WriteByte(mask4 | byte(i>>6&mask44))
		str.WriteByte(mask4 | byte(i&mask44))
	} else if i >= 0x800 {
		str.WriteByte(mask2 | byte(i>>12&mask22))
		str.WriteByte(mask4 | byte(i>>6&mask44))
		str.WriteByte(mask4 | byte(i&mask44))
	} else if i >= 0x80 {
		str.WriteByte(mask1 | byte(i>>6&mask11))
		str.WriteByte(mask4 | byte(i&mask44))
	} else {
		str.WriteByte(byte(i & mask0))
	}
	return str.String()
}
