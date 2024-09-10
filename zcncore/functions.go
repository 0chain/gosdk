package zcncore

import (
	"fmt"
	"github.com/0chain/gosdk/core/transaction"
)

func MinerSCLock(providerId string, providerType Provider, lock uint64) error {
	_, _, _, _, err := transaction.SmartContractTxnValue(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_LOCK,
		InputArgs: &stakePoolRequest{
			ProviderID:   providerId,
			ProviderType: providerType,
		},
	}, lock)

	return err
}

func MinerSCUnlock(providerId string, providerType Provider) error {
	_, _, _, _, err := transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_UNLOCK,
		InputArgs: &stakePoolRequest{
			ProviderID:   providerId,
			ProviderType: providerType,
		},
	})

	return err
}

func MinerSCCollectReward(providerId string, providerType Provider) error {
	_, _, _, _, err := transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.MINERSC_COLLECT_REWARD,
		InputArgs: &scCollectReward{
			ProviderId:   providerId,
			ProviderType: int(providerType),
		},
	})

	return err
}

func MinerSCKill(providerId string, providerType Provider) error {
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
		return fmt.Errorf("kill provider type %v not implimented", providerType)
	}

	_, _, _, _, err := transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      name,
		InputArgs: pr,
	})

	return err
}

func StorageSCCollectReward(providerId string, providerType Provider) error {
	_, _, _, _, err := transaction.SmartContractTxn(StorageSmartContractAddress, transaction.SmartContractTxnData{
		Name: transaction.STORAGESC_COLLECT_REWARD,
		InputArgs: &scCollectReward{
			ProviderId:   providerId,
			ProviderType: int(providerType),
		},
	})

	return err
}

func MinerScUpdateConfig(input interface{}) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_UPDATE_SETTINGS,
		InputArgs: input,
	})

	return err
}

func MinerScUpdateGlobals(input interface{}) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_UPDATE_GLOBALS,
		InputArgs: input,
	})

	return err
}

func StorageScUpdateConfig(input interface{}) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(StorageSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_SETTINGS,
		InputArgs: input,
	})

	return err
}

func AddHardfork(input interface{}) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ADD_HARDFORK,
		InputArgs: input,
	})

	return err
}

func ZCNSCUpdateGlobalConfig(input *MinerSCMinerInfo) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_UPDATE_GLOBAL_CONFIG,
		InputArgs: input,
	})

	return err
}

func MinerSCMinerSettings(input *MinerSCMinerInfo) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_MINER_SETTINGS,
		InputArgs: input,
	})

	return err
}

func MinerSCSharderSettings(input *MinerSCMinerInfo) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_SHARDER_SETTINGS,
		InputArgs: input,
	})

	return err
}

func MinerSCDeleteMiner(input *MinerSCMinerInfo) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_MINER_DELETE,
		InputArgs: input,
	})

	return err
}

func MinerSCDeleteSharder(input *AuthorizerNode) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(MinerSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.MINERSC_SHARDER_DELETE,
		InputArgs: input,
	})

	return err
}

func ZCNSCUpdateAuthorizerConfig(input *AuthorizerNode) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG,
		InputArgs: input,
	})

	return err
}

func ZCNSCAddAuthorizer(input *AddAuthorizerPayload) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_ADD_AUTHORIZER,
		InputArgs: input,
	})

	return err
}

func ZCNSCAuthorizerHealthCheck(input *AuthorizerHealthCheckPayload) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_AUTHORIZER_HEALTH_CHECK,
		InputArgs: input,
	})

	return err
}

func ZCNSCDeleteAuthorizer(input *DeleteAuthorizerPayload) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_DELETE_AUTHORIZER,
		InputArgs: input,
	})

	return err
}

func ZCNSCCollectReward(providerId string, providerType Provider) (err error) {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}

	_, _, _, _, err = transaction.SmartContractTxn(ZCNSCSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.ZCNSC_COLLECT_REWARD,
		InputArgs: pr,
	})

	return err
}
