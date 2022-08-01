package common

import (
	"fmt"
	"path/filepath"
	"strings"
)

func GetPathFields(p string) ([]string, error) {
	if p == "" || p == "/" {
		return nil, nil
	}

	if !filepath.IsAbs(p) {
		return nil, fmt.Errorf("path %s is not absolute", p)
	}

	p = filepath.Clean(p)
	fields := strings.Split(p, "/")
	return fields[1:], nil
}
