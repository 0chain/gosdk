package utils

const (
	Wei    = "wei"
	Kwei   = "kwei"
	Mwei   = "mwei"
	Gwei   = "gwei"
	Szabo  = "szabo"
	Finney = "finney"
	Ether  = "ether"
	Kether = "kether"
	Mether = "mether"
	Gether = "gether"
	Tether = "tether"
)

var units = map[string]string{
	Wei:    "1",
	Kwei:   "1000",
	Mwei:   "1000000",
	Gwei:   "1000000000",
	Szabo:  "1000000000000",
	Finney: "1000000000000000",
	Ether:  "1000000000000000000",
	Kether: "1000000000000000000000",
	Mether: "1000000000000000000000000",
	Gether: "1000000000000000000000000000",
	Tether: "1000000000000000000000000000000",
}
