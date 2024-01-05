package util

import (
	"fmt"
	"os"
	"testing"
)

func TestSymmetricEncryption(t *testing.T) {
	open, _ := os.Open("E:\\IDEA\\JAVA\\xss\\doc\\信手书\\3Workspace\\07实施部署\\产品升级\\版本功能升级\\V2.1.7\\upgrade_V216_V217.zip")
	encryption := Md5Hash(open)
	fmt.Print(encryption)
}
