package util

import (
	"os"

	"go.uber.org/zap"
)

func FileExists(path string) bool {
	zap.S().Debugf("checking if file exists with path: %s", path)
	statInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		zap.S().Errorf("error checking if file exists, unknown state: %s", err)
		return false
	}
	return !statInfo.IsDir()
}

func DirExists(path string) bool {
	zap.S().Debugf("checking if dir exists with path: %s", path)
	statInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		zap.S().Errorf("error checking if dir exists, unknown state: %s", err)
		return false
	}
	return statInfo.IsDir()
}
