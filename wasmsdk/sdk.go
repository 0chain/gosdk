//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"strings"

	"github.com/0chain/gosdk/core/zcncrypto"
)

// Init init sharder/miners ,
func Init(sharders, miners []string) string {

	list := strings.Join(sharders, ",") + ":" + strings.Join(miners, ",")

	return list
}

func TestSign() string {

	sign, err := zcncrypto.SignJsProxy(zcncrypto.Sha3Sum256("data"))
	if err != nil {
		fmt.Println(err)
	}

	return sign
}
