package transaction

// EstimateFee estimates transaction fee
func EstimateFee(txn *Transaction, miners []string, reqPercent float32) (uint64, error) {
	return estimateFee(txn, miners, reqPercent)
}
