package common

import (
	"strings"
)

// Split splits path immediately following the final Separator,
// separating it into a directory and file name component.
// If there is no Separator in path, Split returns an empty dir
// and file set to path.
// The returned values have the property that path = dir+file.
func Split(path string) (dir, file string) {
	if path == "" {
		return "", ""
	}

	if path == "/" {
		return "/", ""
	}

	i := strings.LastIndex(path, "/")

	if i == -1 {
		return "", path
	}

	return string(path[:i]), string(path[i+1:])
}

// Dir returns all but the last element of path, typically the path's directory.
func Dir(path string) string {
	dir, _ := Split(path)

	return dir
}

// Join joins any number of path elements into a single path,
// separating them with slash. Empty elements are ignored.
// The result is Cleaned. However, if the argument
// list is empty or all its elements are empty, Join returns
// an empty string.
func Join(elem ...string) string {
	var items []string
	for _, e := range elem {
		if e != "" {
			if e == "/" && len(items) == 0 {
				items = append(items, "")
				continue
			}

			for _, it := range strings.Split(e, "/") {
				if it != "" {
					items = append(items, it)
				}
			}
		}
	}
	if len(items) == 0 {
		return ""
	}
	return "/" + strings.Join(items, "/")
}

// Clean returns the shortest path name equivalent to path
// by purely lexical processing. It applies the following rules
// iteratively until no further processing can be done:
//
//  1. Replace multiple slash elements with a single one..
//
// The returned path ends in a slash only if it represents a root directory, such as "/".
func Clean(path string) string {
	return Join(path)
}
