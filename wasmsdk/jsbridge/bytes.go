//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"strconv"
	"strings"
)

type Bytes struct {
	Buffer []byte
}

func (bytes *Bytes) UnmarshalJSON(buf []byte) error {

	if len(buf) > 0 {

		src := strings.Trim(string(buf), "\"")

		if len(src) > 0 {
			items := strings.Split(src, ",")

			bytes.Buffer = make([]byte, len(items))

			for k, v := range items {
				it, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return err
				}

				bytes.Buffer[k] = byte(it)
			}

		}

	}

	return nil
}
