package main

import (
	"fmt"

	"github.com/0chain/gosdk/zcncore"
)

func main() {
	fmt.Println("gosdk version: ", zcncore.GetVersion())
}
