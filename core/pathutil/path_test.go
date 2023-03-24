package pathutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplit(t *testing.T) {

	tests := []struct {
		name string
		path string
		dir  string
		file string
	}{
		{
			name: "empty",
			path: "",
			dir:  "",
			file: "",
		},
		{
			name: "only file",
			path: "file",
			dir:  "",
			file: "file",
		},
		{
			name: "only dir",
			path: "/dir/",
			dir:  "/dir",
			file: "",
		},
		{
			name: "only root",
			path: "/",
			dir:  "/",
			file: "",
		},
		{
			name: "single dir and file",
			path: "/file",
			dir:  "/",
			file: "file",
		},
		{
			name: "multiple dirs and file",
			path: "/dir/file",
			dir:  "/dir",
			file: "file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, file := Split(test.path)
			require.Equal(t, test.dir, dir)
			require.Equal(t, test.file, file)
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name string
		elem []string
		path string
	}{
		{name: "empty", elem: []string{"", ""}, path: ""},
		{name: "only slash", elem: []string{"/", "/"}, path: "/"},
		{name: "multiple slashes", elem: []string{"/", "/images"}, path: "/images"},
		{name: "one path without slash", elem: []string{"only_path"}, path: "/only_path"},
		{name: "multiple paths without slash", elem: []string{"path1", "path2"}, path: "/path1/path2"},
		{name: "multiple paths", elem: []string{"path1", "path2", "/path3", "path4/"}, path: "/path1/path2/path3/path4"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := Join(test.elem...)

			require.Equal(t, test.path, path)
		})
	}
}
