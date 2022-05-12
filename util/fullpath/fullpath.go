// From https://github.com/chrislusf/seaweedfs/blob/e41b11b0045376700d4ab047a54a2758a69552ea/weed/util/fullpath.go
// Copyright chrislusf

package fullpath

import (
	"path/filepath"
	"strings"
)

type FullPath string

func NewFullPath(dir, name string) FullPath {
	return FullPath(dir).Child(name)
}

func (fp FullPath) DirAndName() (string, string) {
	dir, name := filepath.Split(string(fp))
	name = strings.ToValidUTF8(name, "?")
	if dir == "/" {
		return dir, name
	}
	if len(dir) < 1 {
		return "/", ""
	}
	return dir[:len(dir)-1], name
}

func (fp FullPath) Name() string {
	_, name := filepath.Split(string(fp))
	name = strings.ToValidUTF8(name, "?")
	return name
}

func (fp FullPath) Child(name string) FullPath {
	dir := string(fp)
	noPrefix := name
	if strings.HasPrefix(name, "/") {
		noPrefix = name[1:]
	}
	if strings.HasSuffix(dir, "/") {
		return FullPath(dir + noPrefix)
	}
	return FullPath(dir + "/" + noPrefix)
}

// Split but skipping the root
func (fp FullPath) Split() []string {
	if fp == "" || fp == "/" {
		return []string{}
	}
	return strings.Split(string(fp)[1:], "/")
}

func Join(names ...string) string {
	return filepath.ToSlash(filepath.Join(names...))
}

func JoinPath(names ...string) FullPath {
	return FullPath(Join(names...))
}
