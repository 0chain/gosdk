// DEPRECATED: This package is deprecated and will be removed in a future release.
package log

import (
	"net/http"
	"os"
	"strings"

	"github.com/0chain/gosdk/core/sys"
)

// HandleFunc returns handle function that writes logs to http.ResponseWriter with provided buffer size.
// Buffered length represented in kilobytes.
func HandleFunc(buffLen int64) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		file, err := os.Open(logName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = file.Close()
		}()

		stat, err := sys.Files.Stat(logName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var (
			writeLen    = buffLen * 1024
			brokenLines = true // flag that describes existence of broken lines
		)
		if writeLen > stat.Size() {
			writeLen = stat.Size()
			brokenLines = false
		}

		buf := make([]byte, writeLen)
		_, err = file.ReadAt(buf, stat.Size()-writeLen)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// cut broken lines if it exist
		if brokenLines {
			lbInd := strings.Index(string(buf), "\n")
			buf = buf[lbInd+1:]
		}

		if _, err := w.Write(buf); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
