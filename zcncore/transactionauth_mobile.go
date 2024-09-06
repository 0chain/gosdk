//go:build mobile
// +build mobile

package zcncore

import (
	"encoding/json"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
)

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string,
	input interface{}, val uint64, feeOpts ...FeeOption) (*transaction.Transaction, error) {
	err := ta.t.createSmartContractTxn(address, methodName, input, val, feeOpts...)
	if err != nil {
		return nil, err
	}
	go func() {
		ta.submitTxn()
	}()
	return ta.t.txn, nil
}

func (ta *TransactionWithAuth) Send(toClientID string, val uint64, desc string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		ta.t.txn.TransactionType = transaction.TxnTypeSend
		ta.t.txn.ToClientID = toClientID
		ta.t.txn.Value = val
		ta.t.txn.TransactionData = string(txnData)
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) VestingAdd(ar VestingAddRequest, value string) error {
	v, err := parseCoinStr(value)
	if err != nil {
		return err
	}

	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, v)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCLock(providerId string, providerType int, lock string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	pr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &pr, lv)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCUnlock(providerId string, providerType int) error {
	pr := &stakePoolRequest{
		ProviderID:   providerId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return err
}

func (ta *TransactionWithAuth) MinerSCCollectReward(providerId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go ta.submitTxn()
	return err
}

func (ta *TransactionWithAuth) StorageSCCollectReward(providerId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return err
}

func (ta *TransactionWithAuth) VestingUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// faucet smart contract

func (ta *TransactionWithAuth) FaucetUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(FaucetSmartContractAddress,
		transaction.FAUCETSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateGlobals(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) StorageScUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateGlobalConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress,
		transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) GetVerifyConfirmationStatus() int {
	return ta.t.GetVerifyConfirmationStatus()
}

func (ta *TransactionWithAuth) MinerSCMinerSettings(info MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCSharderSettings(info MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteMiner(info MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteSharder(info MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateAuthorizerConfig(ip AuthorizerNode) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAddAuthorizer(ip AddAuthorizerPayload) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_ADD_AUTHORIZER, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAuthorizerHealthCheck(ip *AuthorizerHealthCheckPayload) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_AUTHORIZER_HEALTH_CHECK, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.t.setNonceAndSubmit()
	return
}
