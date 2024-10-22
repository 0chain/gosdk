package zcncore

import (
	"fmt"
	"github.com/0chain/gosdk/core/transaction"
)

// AuthorizerNode represents an authorizer node in the network
type AuthorizerNode struct {
	ID     string            `json:"id"`
	URL    string            `json:"url"`
	Config *AuthorizerConfig `json:"config"`
}

type scCollectReward struct {
	ProviderId   string `json:"provider_id"`
	ProviderType int    `json:"provider_type"`
}

type MinerSCDelegatePool struct {
	Settings StakePoolSettings `json:"settings"`
}

// SimpleMiner represents a node in the network, miner or sharder.
type SimpleMiner struct {
	ID string `json:"id"`
}

// MinerSCMinerInfo interface for miner/sharder info functions on miner smart contract.
type MinerSCMinerInfo struct {
	SimpleMiner         `json:"simple_miner"`
	MinerSCDelegatePool `json:"stake_pool"`
}

func MinerSCLock(providerId string, providerType Provider, lock uint64, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxnValue(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_LOCK,
		InputArgs: &stakePoolRequest{
			ProviderID:   providerId,
			ProviderType: providerType,
		},
	}, lock, client...)

}

func MinerSCUnlock(providerId string, providerType Provider, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_UNLOCK,
		InputArgs: &stakePoolRequest{
			ProviderID:   providerId,
			ProviderType: providerType,
		},
	}, client...)
}

func MinerSCCollectReward(providerId string, providerType Provider, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_COLLECT_REWARD,
		InputArgs: &scCollectReward{
			ProviderId:   providerId,
			ProviderType: int(providerType),
		},
	}, client...)

}

func MinerSCKill(providerId string, providerType Provider, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	var name string
	switch providerType {
	case ProviderMiner:
		name = transaction.MINERSC_KILL_MINER
	case ProviderSharder:
		name = transaction.MINERSC_KILL_SHARDER
	default:
		return "", "", -1, &transaction.Transaction{}, fmt.Errorf("kill provider type %v not implimented", providerType)
	}

	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      name,
		InputArgs: pr,
	}, client...)

}

func StorageSCCollectReward(providerId string, providerType Provider, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(StorageSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.STORAGESC_COLLECT_REWARD,
		InputArgs: &scCollectReward{
			ProviderId:   providerId,
			ProviderType: int(providerType),
		},
	}, client...)

}

func MinerScUpdateConfig(input interface{}, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_UPDATE_SETTINGS,
		InputArgs: input,
	}, client...)

}

func MinerScUpdateGlobals(input interface{}, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_UPDATE_GLOBALS,
		InputArgs: input,
	}, client...)

}

func StorageScUpdateConfig(input interface{}, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(StorageSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_SETTINGS,
		InputArgs: input,
	}, client...)

}

func AddHardfork(input interface{}, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ADD_HARDFORK,
		InputArgs: input,
	}, client...)

}

func ZCNSCUpdateGlobalConfig(input *InputMap, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_UPDATE_GLOBAL_CONFIG,
		InputArgs: input,
	}, client...)

}

func MinerSCMinerSettings(input *MinerSCMinerInfo, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_MINER_SETTINGS,
		InputArgs: input,
	}, client...)

}

func MinerSCSharderSettings(input *MinerSCMinerInfo, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_SHARDER_SETTINGS,
		InputArgs: input,
	}, client...)

}

func ZCNSCUpdateAuthorizerConfig(input *AuthorizerNode, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG,
		InputArgs: input,
	}, client...)

}

func ZCNSCAddAuthorizer(input *AddAuthorizerPayload, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_ADD_AUTHORIZER,
		InputArgs: input,
	}, client...)

}

func ZCNSCAuthorizerHealthCheck(input *AuthorizerHealthCheckPayload, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_AUTHORIZER_HEALTH_CHECK,
		InputArgs: input,
	}, client...)

}

func ZCNSCDeleteAuthorizer(input *DeleteAuthorizerPayload, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_DELETE_AUTHORIZER,
		InputArgs: input,
	}, client...)

}

func ZCNSCCollectReward(providerId string, providerType Provider, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}

	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_COLLECT_REWARD,
		InputArgs: pr,
	}, client...)
}

type SendTxnData struct {
	Note string `json:"note"`
}

func Send(toClientID string, tokens uint64, desc string, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	if len(client) == 0 {
		client = append(client, "")
		client = append(client, toClientID)
	} else {
		client = append(client, toClientID)
	}
	return transaction.SmartContractTxnValue(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      "transfer",
		InputArgs: SendTxnData{Note: desc},
	}, tokens, client...)
}

func Faucet(tokens uint64, input string, client ...string) (hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return transaction.SmartContractTxnValue(FaucetSmartContractAddress, transaction.SmartContractTxnData{
		Name:      "pour",
		InputArgs: input,
	}, tokens, client...)
}
