//go:build !mobile
// +build !mobile

package zcncore

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string, input interface{}, val int64) error {
	err := ta.t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}
