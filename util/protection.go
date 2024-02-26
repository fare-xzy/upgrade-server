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

func Md5HashByByte(origin []byte) string {

	// 合并数据和盐值
	dataWithSalt := append(origin, salt...)
	// 创建MD5哈希器
	hasher := md5.New()

	// 将字符串转换为字节数组并计算哈希值
	hasher.Write(dataWithSalt)
	sum := hasher.Sum(nil)

	return hex.EncodeToString(sum)
}
