package zcncore

import "github.com/0chain/gosdk/core/transaction"

func StorageScUpdateConfig(input string) (err error) {
	_, _, _, _, err = transaction.SmartContractTxn(StorageSmartContractAddress, transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_SETTINGS,
		InputArgs: input,
	})

	return err
}
