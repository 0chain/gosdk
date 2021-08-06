module github.com/0chain/gosdk

require (
	github.com/aristanetworks/goarista v0.0.0-20190912214011-b54698eaaca6 // indirect
	github.com/elastic/gosigar v0.10.5 // indirect
	github.com/ethereum/go-ethereum v1.10.4
	github.com/h2non/filetype v1.0.9
	github.com/herumi/bls-go-binary v1.0.0
	github.com/klauspost/reedsolomon v1.9.11
	github.com/miguelmota/go-ethereum-hdwallet v0.1.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570 // indirect
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tyler-smith/go-bip39 v1.1.0
	go.dedis.ch/kyber/v3 v3.0.5
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
)

go 1.13

replace github.com/ethereum/go-ethereum => github.com/certifaction/go-ethereum v1.10.3-wasm
