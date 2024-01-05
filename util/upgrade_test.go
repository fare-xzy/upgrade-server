package util

import "testing"

func TestConvertOctonaryUtf8(t *testing.T) {
	str := "upgrade/application#\\346\\234\\215\\345\\212\\241\\347\\211\\210\\346\\234\\254\\345\\215\\207\\347\\272\\247/recover.sh"
	ConvertOctonaryUtf8(str)
}
