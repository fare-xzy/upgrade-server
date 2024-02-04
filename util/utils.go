package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

func GetWorkPath() string {
	pwd, _ := os.Getwd()
	filePath := filepath.Join(pwd, "tmp", GetTag())
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.MkdirAll(filePath, 0755)
	}
	return filePath
}

func GetPwd() string {
	pwd, _ := os.Getwd()
	return pwd
}

func GetTag() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func CloseAndWait(stop, closed chan bool, timeout time.Duration) error {
	select {
	case _, ok := <-stop:
		if !ok {
			return nil
		}
	default:
	}

	close(stop)

	select {
	case <-closed:
		return nil
	case <-time.After(timeout):
		return Err("Wait for closed timeout")
	}
}

func Sprintf(f string, args ...interface{}) string {
	return fmt.Sprintf(f, args...)
}

func Err(f string, args ...interface{}) error {
	return errors.New(Sprintf(f, args...))
}

func GetFilenameExtension(path string) (ext string) {
	var extIndex int
	if path == "" {
		return
	}
	if extIndex = strings.LastIndex(path, "."); extIndex == -1 {
		return
	}
	folderIndex := strings.LastIndex(path, "/")
	if folderIndex > extIndex {
		return
	}
	runes := []rune(path)
	ext = string(runes[extIndex+1:])
	return
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func U16StrToUTF8Str(u16Str string) (string, error) {
	reg := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)
	s := reg.ReplaceAllStringFunc(u16Str, func(s string) string {
		s = s[2:] // strip \\u
		u16, err := strconv.ParseUint(s, 16, 16)
		if err != nil {
			return ""
		}
		return string(utf16.Decode([]uint16{uint16(u16)}))
	})
	return s, nil
}
