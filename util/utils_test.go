package util

import (
	"fmt"
	"testing"
)

func TestU16StrToUTF8Str(t *testing.T) {
	str, err := U16StrToUTF8Str("# Default 0.5M\\uFF0C1\\u3001\\u5355\\u4E2A\\u5370\\u7AE0\\u5927\\u5C0F<=0.5M\\uFF0C2\\u3001\\u6279\\u91CF\\u4E0A\\u4F20\\u5370\\u7AE0\\u8981\\u6C42\\u5355\\u4E2A\\u5370\\u7AE0\\u5927\\u5C0F<=0.5M\\uFF0C")
	if err != nil {
		return
	}
	fmt.Println(str)
}
