//go:build !mobile
// +build !mobile

package zcncore

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/transaction"
	"math"
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
	txnData, err := json.Marshal(transaction.SmartContractTxnData{Name: "transfer", InputArgs: SendTxnData{Note: desc}})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}

	clientNode, err := client.GetNode()
	if err != nil {
		return err
	}

	ta.t.txn.TransactionType = transaction.TxnTypeSend
	ta.t.txn.ToClientID = toClientID
	ta.t.txn.Value = val
	ta.t.txn.TransactionData = string(txnData)
	if ta.t.txn.TransactionFee == 0 {
		fee, err := transaction.EstimateFee(ta.t.txn, clientNode.Network().Miners, 0.2)
		if err != nil {
			return err
		}
		ta.t.txn.TransactionFee = fee
	}

	go func() {
		ta.submitTxn()
	}()

	return nil
}

func (ta *TransactionWithAuth) MinerSCLock(providerId string, providerType Provider, lock uint64) error {
	if lock > math.MaxInt64 {
		return errors.New("invalid_lock", "int64 overflow on lock value")
	}

	pr := &stakePoolRequest{
		ProviderID:   providerId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, pr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCUnlock(providerId string, providerType Provider) error {
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

func (ta *TransactionWithAuth) MinerSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
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

func (ta *TransactionWithAuth) MinerSCKill(providerId string, providerType Provider) error {
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
	err := ta.t.createSmartContractTxn(MinerSmartContractAddress, name, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go ta.submitTxn()
	return err
}

func (ta *TransactionWithAuth) StorageSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
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

// faucet smart contract

func (ta *TransactionWithAuth) FaucetUpdateConfig(ip *InputMap) (err error) {
	err = ta.t.createSmartContractTxn(FaucetSmartContractAddress,
		transaction.FAUCETSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateConfig(ip *InputMap) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateGlobals(ip *InputMap) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) StorageScUpdateConfig(ip *InputMap) (err error) {
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (t *TransactionWithAuth) AddHardfork(ip *InputMap) (err error) {
	err = t.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.ADD_HARDFORK, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { t.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateGlobalConfig(ip *InputMap) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) GetVerifyConfirmationStatus() ConfirmationStatus {
	return ta.t.GetVerifyConfirmationStatus() //nolint
}

func (ta *TransactionWithAuth) MinerSCMinerSettings(info *MinerSCMinerInfo) (
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

func (ta *TransactionWithAuth) MinerSCSharderSettings(info *MinerSCMinerInfo) (
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

func (ta *TransactionWithAuth) MinerSCDeleteMiner(info *MinerSCMinerInfo) (
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

func (ta *TransactionWithAuth) MinerSCDeleteSharder(info *MinerSCMinerInfo) (
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

func (ta *TransactionWithAuth) ZCNSCUpdateAuthorizerConfig(ip *AuthorizerNode) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAddAuthorizer(ip *AddAuthorizerPayload) (err error) {
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

func (ta *TransactionWithAuth) ZCNSCDeleteAuthorizer(ip *DeleteAuthorizerPayload) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_DELETE_AUTHORIZER, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	err := ta.t.createSmartContractTxn(ZCNSCSmartContractAddress,
		transaction.ZCNSC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.t.setNonceAndSubmit() }()
	return err
}
