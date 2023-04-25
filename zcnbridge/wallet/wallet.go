package wallet

import (
	//"github.com/0chain/gosdk/zcnbridge/log"
	"os"

	"github.com/0chain/gosdk/core/logger"
)

const (
	ZCNSCSmartContractAddress = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0"
	MintFunc                  = "mint"
	BurnFunc                  = "burn"
	BurnWzcnTicketPath        = "/v1/ether/burnticket/"
	BurnWzcnBurnEventsPath    = "/v1/ether/burnevents/"
	BurnNativeTicketPath      = "/v1/0chain/burnticket/"
)

var Logger logger.Logger
var defaultLogLevel = logger.DEBUG

func init() {
	Logger.Init(defaultLogLevel, "zcnbridge-wallet-sdk")

	Logger.SetLevel(logger.DEBUG)
	f, err := os.OpenFile("bridge.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	Logger.SetLogFile(f, true)
}
