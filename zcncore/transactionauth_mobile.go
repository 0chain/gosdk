//go:build mobile
// +build mobile

package zcncore

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string, input string, val int64) error {
	err := ta.t.createSmartContractTxnWithJSON(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}
