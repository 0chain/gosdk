// Provides the data structures and methods to work with the common data structures.
package common

import (
	"fmt"
	"path"
	"strings"

	"github.com/valyala/bytebufferpool"
)

var MemPool bytebufferpool.Pool

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
