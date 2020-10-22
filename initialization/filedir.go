package initialization

import (
	"os"

	t "github.com/gaku/BiuBiun/tool"
)

func InitFileDir() (err error) {
	//創建GinLogDir
	logDirPath := os.Getenv("GinDirPath")
	exits, err := t.PathExists(logDirPath)
	if err != nil {
		return
	}
	if !exits {
		if err = os.MkdirAll(logDirPath, os.ModePerm); err != nil {
			return
		}
	}
	//創建UserFileDir
	UDirPath := os.Getenv("MemberFileDirPath")
	exists, err := t.PathExists(UDirPath)
	if err != nil {
		return
	}
	if !exists {
		if err = os.MkdirAll(UDirPath, os.ModePerm); err != nil {
			return
		}
	}
	//創建ShopFileDir
	SDirPath := os.Getenv("ShopFileDirPath")
	if err != nil {
		return
	}
	if !exists {
		if err = os.MkdirAll(SDirPath, os.ModePerm); err != nil {
			return
		}
	}
	return
}
