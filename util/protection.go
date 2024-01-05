package util

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"mime/multipart"
)

const salt = "JdZj7WuQmFAx7egWrEUxAmZMvqDLJC8Y"

func Md5Hash(origin multipart.File) string {

	hasher := md5.New()
	buffer := make([]byte, 4096) // 4KB 缓冲区
	for {
		n, err := origin.Read(buffer)
		if err == io.EOF {
			// 读取完文件后追加盐值
			hasher.Write([]byte(salt))
			break
		}

		// 更新 MD5 哈希
		hasher.Write(buffer[:n])
	}

	// 计算 MD5 值
	md5Sum := hasher.Sum(nil)

	return hex.EncodeToString(md5Sum)
}
