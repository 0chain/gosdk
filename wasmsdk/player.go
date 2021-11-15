// +build js,wasm

package main

import (
	"fmt"
	"net/http"
)

type M3u8Server struct {
}

func (s *M3u8Server) Start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, wasm")
	})

	fmt.Println("Start m3u8 server")
	http.ListenAndServe(":12345", nil) //nolint
}
