//go:build !mobile
// +build !mobile

package zcncore

// TransactionScheme implements few methods for block chain.
//
// Note: to be buildable on MacOSX all arguments should have names.
type TransactionScheme interface {
	// SetTransactionCallback implements storing the callback
	// used to call after the transaction or verification is completed
	SetTransactionCallback(cb TransactionCallback) error
	// Send implements sending token to a given clientid
	Send(toClientID string, val int64, desc string) error
	// StoreData implements store the data to blockchain
	StoreData(data string) error
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val int64) error
	// ExecuteFaucetSCWallet implements the `Faucet Smart contract` for a given wallet
	ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error
	// GetTransactionHash implements retrieval of hash of the submitted transaction
	GetTransactionHash() string
	//RegisterMultiSig registers a group wallet and subwallets with MultisigSC
	RegisterMultiSig(walletstr, mswallet string) error
	// SetTransactionHash implements verify a previous transaction status
	SetTransactionHash(hash string) error
	// SetTransactionFee implements method to set the transaction fee
	SetTransactionFee(txnFee int64) error
	// SetTransactionNonce implements method to set the transaction nonce
	SetTransactionNonce(txnNonce int64) error
	// Verify implements verify the transaction
	Verify() error
	// GetVerifyConfirmationStatus implements the verification status from sharders
	GetVerifyConfirmationStatus() ConfirmationStatus
	// GetVerifyOutput implements the verification output from sharders
	GetVerifyOutput() string
	// GetTransactionError implements error string in case of transaction failure
	GetTransactionError() string
	// GetVerifyError implements error string in case of verify failure error
	GetVerifyError() string
	// GetTransactionNonce returns nonce
	GetTransactionNonce() int64

	// Output of transaction.
	Output() []byte

	// Hash Transaction status regardless of status
	Hash() string

	// Vesting SC

	VestingTrigger(poolID string) error
	VestingStop(sr *VestingStopRequest) error
	VestingUnlock(poolID string) error
	VestingAdd(ar *VestingAddRequest, value int64) error
	VestingDelete(poolID string) error
	VestingUpdateConfig(*InputMap) error

	// Miner SC

	MinerSCCollectReward(string, string, Provider) error
	MinerSCMinerSettings(*MinerSCMinerInfo) error
	MinerSCSharderSettings(*MinerSCMinerInfo) error
	MinerSCLock(minerID string, lock int64) error
	MinerSCUnlock(minerID, poolID string) error
	MinerScUpdateConfig(*InputMap) error
	MinerScUpdateGlobals(*InputMap) error
	MinerSCDeleteMiner(*MinerSCMinerInfo) error
	MinerSCDeleteSharder(*MinerSCMinerInfo) error

	// Storage SC

	StorageSCCollectReward(string, string, Provider) error
	FinalizeAllocation(allocID string, fee int64) error
	CancelAllocation(allocID string, fee int64) error
	CreateAllocation(car *CreateAllocationRequest, lock, fee int64) error //
	CreateReadPool(fee int64) error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock, fee int64) error
	ReadPoolUnlock(poolID string, fee int64) error
	StakePoolLock(blobberID string, lock, fee int64) error
	StakePoolUnlock(blobberID string, poolID string, fee int64) error
	UpdateBlobberSettings(blobber *Blobber, fee int64) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock, fee int64) error
	WritePoolLock(allocID string, blobberID string, duration int64, lock, fee int64) error
	WritePoolUnlock(poolID string, fee int64) error
	StorageScUpdateConfig(*InputMap) error

	// Faucet

	FaucetUpdateConfig(*InputMap) error

	// ZCNSC Common transactions

	// ZCNSCUpdateGlobalConfig updates global config
	ZCNSCUpdateGlobalConfig(*InputMap) error
	// ZCNSCUpdateAuthorizerConfig updates authorizer config by ID
	ZCNSCUpdateAuthorizerConfig(*AuthorizerNode) error
	// ZCNSCAddAuthorizer adds authorizer
	ZCNSCAddAuthorizer(*AddAuthorizerPayload) error
}
