package wallet

import (
	"github.com/0chain/gosdk/core/logger"
)

const (
	ZCNSCSmartContractAddress = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0"
	AddAuthorizerFunc         = "AddAuthorizer"
	DeleteAuthorizerFunc      = "DeleteAuthorizer"
	MintFunc                  = "mint"
	BurnFunc                  = "burn"
	ConsensusThresh           = float64(70.0)
	BurnTicketPath            = "/v1/ether/burnticket/get"
)

var ClientID string
var Logger logger.Logger
