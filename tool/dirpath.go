package tool

import (
	"os"

	"github.com/pkg/errors"
)

func PathExists(Path string) (exists bool, err error) {
	_, err = os.Stat(Path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.WithStack(err)
}
