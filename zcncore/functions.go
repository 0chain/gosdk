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
