package regexp

import (
	"regexp"
	"testing"
)

// go test -v -run TestReplaceUnicode
func TestReplaceUnicode(t *testing.T) {
	// 这种字符串只能在源码中使用,包含中文使用\u这种格式
	s0 := "^[\u4e00-\u9fa5\\w]{1,20}$"
	r0 := regexp.MustCompile(s0)

	// 这种是外部传入\u格式字符,不经过转换regexp库会直接报错
	s1 := ReplaceUnicode([]byte(`^[\u4e00-\u9fa5\w]{1,20}$`))
	r1 := regexp.MustCompile(s1)

	if s0 != s1 {
		t.Fatal("s0 != s1")
	}

	if r0.String() != r1.String() {
		t.Fatal("r0 != r1")
	}
}
