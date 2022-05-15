module example

replace github.com/0chain/gosdk => ../../

//replace 0chain.net/0chain/authorizer => ../example/authorizer

require (
	github.com/0chain/gosdk v0.0.0
	go.uber.org/zap v1.21.0
)

go 1.17
