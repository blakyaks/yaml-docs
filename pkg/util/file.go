package util

import (
	"os"
	"path/filepath"
)

func IsRelativePath(filePath string) bool {
	return (filePath[0] == '.') && filepath.Base(filePath) != filePath
}

func IsBaseFilename(filePath string) bool {
	return filepath.Base(filePath) == filePath
}

func GetAbsolutePath(filePath string) string {
	if !filepath.IsAbs(filePath) {
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, filePath)
	} else {
		return filePath
	}
}

func GetParentDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

func GetBaseFilename(filePath string) string {
	return filepath.Base(filePath)
}
