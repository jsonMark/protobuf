package utils

import (
	"os"
)

// FileOrDirExists 判断文件夹是否存在
func FileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// CreateNoExistsDir 创建文件夹,存在跳过
func CreateNoExistsDir(dirPath string) error {
	exist := FileOrDirExists(dirPath)
	if !exist {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
