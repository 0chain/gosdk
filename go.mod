module github.com/0chain/gosdk

go 1.16

require (
	github.com/0chain/errors v1.0.2
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/dgraph-io/badger/v3 v3.2103.1
	github.com/didip/tollbooth v4.0.2+incompatible
	github.com/ethereum/go-ethereum v1.10.8
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/h2non/filetype v1.0.9
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/herumi/bls-go-binary v1.0.0
	github.com/klauspost/cpuid/v2 v2.0.4 // indirect
	github.com/klauspost/reedsolomon v1.9.11
	github.com/kr/text v0.2.0 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7
	github.com/magma/augmented-networks/accounting/protos v0.1.1
	github.com/mitchellh/mapstructure v1.4.2
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/rogpeppe/go-internal v1.3.0
	github.com/spf13/viper v1.9.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tyler-smith/go-bip39 v1.1.0
	go.dedis.ch/kyber/v3 v3.0.5
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	google.golang.org/grpc v1.40.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/ethereum/go-ethereum => github.com/certifaction/go-ethereum v1.10.3-wasm
