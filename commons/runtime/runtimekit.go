package runtime

import (
	"os"
	"path/filepath"
	"strings"
)

//返回工作目录
func GetWorkDir() string {
	wd, _ := os.Getwd()
	return wd
}

func GetBinDir() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return strings.Replace(dir, "\\", "/", -1)
}
