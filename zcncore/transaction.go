//go:build !gomobile
// +build !gomobile

package zcncore

func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val int64) error {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}
