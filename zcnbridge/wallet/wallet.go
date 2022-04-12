package wallet

import (
	//"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/core/logger"
)

const (
	ZCNSCSmartContractAddress = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0"
	MintFunc                  = "mint"
	BurnFunc                  = "burn"
	BurnWzcnTicketPath        = "/v1/ether/burnticket/get"
	BurnNativeTicketPath      = "/v1/0chain/burnticket/get"
)

var Logger logger.Logger
var defaultLogLevel = logger.DEBUG

func init() {
	Logger.Init(defaultLogLevel, "0chain-zcnbridge-sdk")
}
