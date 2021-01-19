package bls

import (
  "fmt"
  "github.com/0chain/gosdk/miracl"
)

// Copied directly from herumi's source code.
// Sign: <https://github.com/herumi/bls-go-binary/blob/ef6a150a928bddb19cee55aec5c80585528d9a96/bls/bls.go#L449>
// blsSignature: <https://github.com/herumi/bls/blob/1b48de51f4f76deb204d108f6126c1507623f739/include/bls/bls.h#L68>
// mclBnG1: <https://github.com/herumi/mcl/blob/0114a3029f74829e79dc51de6dfb28f5da580632/include/mcl/bn.h#L96>
type Sign struct {
  v BN254.ECP
}

func Init() {
  fmt.Println("hello world")
}
