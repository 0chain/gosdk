//go:build mobile
// +build mobile

package zcncore

func (t *Transaction) ExecuteSmartContract(address, methodName string, input string, val int64) error {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}
