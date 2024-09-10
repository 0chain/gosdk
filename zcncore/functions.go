package zcncore

import (
	"fmt"
	"github.com/0chain/gosdk/core/transaction"
)

func MinerSCLock(providerId string, providerType Provider, lock uint64) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxnValue(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_LOCK,
		InputArgs: &stakePoolRequest{
			ProviderID:   providerId,
			ProviderType: providerType,
		},
	}, lock)

}

func MinerSCUnlock(providerId string, providerType Provider) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_UNLOCK,
		InputArgs: &stakePoolRequest{
			ProviderID:   providerId,
			ProviderType: providerType,
		},
	})

}

func MinerSCCollectReward(providerId string, providerType Provider) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_COLLECT_REWARD,
		InputArgs: &scCollectReward{
			ProviderId:   providerId,
			ProviderType: int(providerType),
		},
	})

}

func MinerSCKill(providerId string, providerType Provider) (string, string, int64, *transaction.Transaction, error) {
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
	})

}

func StorageSCCollectReward(providerId string, providerType Provider) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(StorageSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.STORAGESC_COLLECT_REWARD,
		InputArgs: &scCollectReward{
			ProviderId:   providerId,
			ProviderType: int(providerType),
		},
	})

}

func MinerScUpdateConfig(input interface{}) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_UPDATE_SETTINGS,
		InputArgs: input,
	})

}

func MinerScUpdateGlobals(input interface{}) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_UPDATE_GLOBALS,
		InputArgs: input,
	})

}

func StorageScUpdateConfig(input interface{}) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(StorageSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_SETTINGS,
		InputArgs: input,
	})

}

func AddHardfork(input interface{}) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ADD_HARDFORK,
		InputArgs: input,
	})

}

func ZCNSCUpdateGlobalConfig(input *MinerSCMinerInfo) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_UPDATE_GLOBAL_CONFIG,
		InputArgs: input,
	})

}

func MinerSCMinerSettings(input *MinerSCMinerInfo) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_MINER_SETTINGS,
		InputArgs: input,
	})

}

func MinerSCSharderSettings(input *MinerSCMinerInfo) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_SHARDER_SETTINGS,
		InputArgs: input,
	})

}

func MinerSCDeleteMiner(input *MinerSCMinerInfo) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_MINER_DELETE,
		InputArgs: input,
	})

}

func MinerSCDeleteSharder(input *AuthorizerNode) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_SHARDER_DELETE,
		InputArgs: input,
	})

}

func ZCNSCUpdateAuthorizerConfig(input *AuthorizerNode) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG,
		InputArgs: input,
	})

}

func ZCNSCAddAuthorizer(input *AddAuthorizerPayload) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_ADD_AUTHORIZER,
		InputArgs: input,
	})

}

func ZCNSCAuthorizerHealthCheck(input *AuthorizerHealthCheckPayload) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_AUTHORIZER_HEALTH_CHECK,
		InputArgs: input,
	})

}

func ZCNSCDeleteAuthorizer(input *DeleteAuthorizerPayload) (string, string, int64, *transaction.Transaction, error) {
	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_DELETE_AUTHORIZER,
		InputArgs: input,
	})

}

func ZCNSCCollectReward(providerId string, providerType Provider) (string, string, int64, *transaction.Transaction, error) {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}

	return transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_COLLECT_REWARD,
		InputArgs: pr,
	})

}
