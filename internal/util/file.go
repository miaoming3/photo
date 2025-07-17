package util

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetFileMD5 计算文件内容的MD5值
func GetFileMD5(file io.Reader) (string, error) {
	hm := md5.New()
	if _, err := io.Copy(hm, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hm.Sum(nil)), nil
}

// CheckFileExists 检查文件是否存在
func CheckFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureDir 创建目录（如果不存在）
func EnsureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// GetFileContentType 获取文件MIME类型
func GetFileContentType(file io.Reader) (string, error) {
	// 读取前512字节来检测文件类型
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// 使用http.DetectContentType检测MIME类型
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

// SanitizeFileName 清理文件名，移除特殊字符
func SanitizeFileName(name string) string {
	// 定义不允许的字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "<", ">", "|", "\""}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name
}

// SaveUploadedFile 保存上传的文件到指定路径
func SaveUploadedFile(file io.Reader, dstPath string) error {
	// 创建目标目录
	if err := EnsureDir(filepath.Dir(dstPath)); err != nil {
		return err
	}

	// 创建目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, file)
	return err
}

// GetFileSize 获取文件大小
func GetFileSize(file io.Reader) (int64, error) {
	// 如果文件实现了Size方法（如*os.File）
	if f, ok := file.(interface{ Size() int64 }); ok {
		return f.Size(), nil
	}

	// 否则通过复制到null设备来计算大小
	var size int64
	_, err := io.Copy(io.Discard, io.TeeReader(file, &sizeCounter{&size}))
	return size, err
}

// sizeCounter 用于计算读取的字节数
type sizeCounter struct {
	size *int64
}

func (c *sizeCounter) Write(p []byte) (int, error) {
	*c.size += int64(len(p))
	return len(p), nil
}

// IsAllowedFileType 检查文件类型是否在允许列表中
func IsAllowedFileType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if strings.EqualFold(allowed, contentType) || strings.HasPrefix(strings.ToLower(contentType), strings.ToLower(allowed)) {
			return true
		}
	}
	return false
}
