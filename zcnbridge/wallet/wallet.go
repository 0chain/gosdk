package wallet

import (
	"github.com/0chain/gosdk/core/logger"
	"gopkg.in/natefinch/lumberjack.v2"
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
	ioWriter := &lumberjack.Logger{
		Filename:   "bridge.log",
		MaxSize:    100, // MB
		MaxBackups: 5,   // number of backups
		MaxAge:     28,  //days
		LocalTime:  false,
		Compress:   false, // disabled by default
	}
	Logger.SetLogFile(ioWriter, true)
}
