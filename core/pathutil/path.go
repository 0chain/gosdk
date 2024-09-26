// Provides utility functions for working with file paths.
package pathutil

import (
	"strings"
)

// Split splits path immediately following the final Separator,
// separating it into a directory and file name component.
// If there is no Separator in path, Split returns an empty dir
// and file set to path.
// The returned values have the property that path = dir+file.
//   - path is the path to be split.
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
//   - path is the path to be split.
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
	var hasElements bool
	for _, e := range elem {
		if e != "" {
			hasElements = true
			for _, it := range strings.Split(e, "/") {
				if it != "" {
					items = append(items, it)
				}
			}
		}
	}
	if !hasElements {
		return ""
	}

	return "/" + strings.Join(items, "/")

}
