//go:build mobile
// +build mobile

package zcncore

type SmartContractExecutor interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input string, val uint64) error
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input string, val uint64) error {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string, input string, val uint64) error {
	err := ta.t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}
