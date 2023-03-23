package common

import (
	"fmt"
	"path"
	"strings"
)

func GetPathDir(path string) string {

	i := strings.LastIndex(path, "/")

	return string(path[:i])

}

func GetPathFields(p string) ([]string, error) {
	if p == "" || p == "/" {
		return nil, nil
	}

	if !path.IsAbs(p) {
		return nil, fmt.Errorf("path %s is not absolute", p)
	}

	p = path.Clean(p)
	fields := strings.Split(p, "/")
	return fields[1:], nil
}
