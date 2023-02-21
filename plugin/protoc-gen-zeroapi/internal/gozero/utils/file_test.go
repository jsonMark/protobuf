package utils

import "testing"

func TestCreateDir(t *testing.T) {

	dirPath := "my_dir/test"

	err := CreateNoExistsDir(dirPath)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Log("success")
}

func TestFileOrDirExists(t *testing.T) {

	dirPath := "tools.go"
	if exists := FileOrDirExists(dirPath); !exists {
		t.Errorf("error")
		return
	}

	t.Log("success")
}
