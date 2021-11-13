module sample

replace (
	github.com/0chain/gosdk => ../../
)

require (
	github.com/0chain/gosdk v0.0.0
	github.com/spf13/viper v1.7.0
	go.uber.org/zap v1.15.0
)

go 1.16