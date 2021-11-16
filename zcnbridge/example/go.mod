module sample

replace github.com/0chain/gosdk => ../../

require (
	github.com/0chain/gosdk v0.0.0
	go.uber.org/zap v1.19.1
)

go 1.16
