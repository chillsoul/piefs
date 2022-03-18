package util

import "os"

func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
func PathMustExists(path string) {
	exists := PathExist(path)
	if !exists {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			panic("Path Must exists: " + err.Error())
		}
	}
}
