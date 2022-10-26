package zcncore

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zcncore"
)

const (
	GET_LOCK_CONFIG   = `/v1/screst/` + InterestPoolSmartContractAddress + `/getLockConfig`
	GET_LOCKED_TOKENS = `/v1/screst/` + InterestPoolSmartContractAddress + `/getPoolsStats?client_id=`

	// inerest pool SC

	INTERESTPOOLSC_PFX        = `/v1/screst/` + InterestPoolSmartContractAddress
	GET_INTERESTPOOLSC_CONFIG = INTERESTPOOLSC_PFX + `/getConfig`

	// storage SC

	STORAGESC_PFX = "/v1/screst/" + StorageSmartContractAddress

	STORAGESC_GET_WRITE_POOL_INFO = STORAGESC_PFX + "/getWritePoolStat"
)

const (
	StorageSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`

	InterestPoolSmartContractAddress = `cf8d0df9bd8cc637a4ff4e792ffe3686da6220c45f0e1103baa609f3f1751ef4`
)

const (
	OpGetTokenLockConfig int = iota
	OpGetLockedTokens
	OpGetUserPools
	OpGetUserPoolDetail
	// storage SC ops
	OpStorageSCGetConfig
	OpStorageSCGetChallengePoolInfo
	OpStorageSCGetAllocation
	OpStorageSCGetAllocations
	OpStorageSCGetReadPoolInfo
	OpStorageSCGetStakePoolInfo
	OpStorageSCGetBlobbers
	OpStorageSCGetBlobber
	OpStorageSCGetTransactions
	OpStorageSCGetWritePoolInfo
)

// GetInfoCallback needs to be implemented by the caller of GetLockTokenConfig() and GetLockedTokens()
type GetInfoCallback interface {
	// OnInfoAvailable will be called when GetLockTokenConfig is complete
	// if status == StatusSuccess then info is valid
	// is status != StatusSuccess then err will give the reason
	OnInfoAvailable(op int, status int, info string, err string)
}

// Init inializes the SDK with miner, sharder and signature scheme provided in
// configuration provided in JSON format
// It is used for 0proxy, 0box, 0explorer, andorid, ios : walletJSON is ChainConfig
//	 {
//    "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
//		"signature_scheme" : "bls0chain",
//		"block_worker" : "http://localhost/dns",
// 		"min_submit" : 50,
//		"min_confirmation" : 50,
//		"confirmation_chain_length" : 3,
//		"num_keys" : 1,
//		"eth_node" : "https://ropsten.infura.io/v3/xxxxxxxxxxxxxxx"
//	 }
func Init(chainConfigJSON string) error {
	var chain zcncore.ChainConfig
	err := json.Unmarshal([]byte(chainConfigJSON), &chain)
	if err == nil {
		// Check signature scheme is supported
		if chain.SignatureScheme != "ed25519" && chain.SignatureScheme != "bls0chain" {
			return errors.New("", "invalid/unsupported signature scheme")
		}

		return zcncore.InitZCNSDK(chain.BlockWorker,
			chain.SignatureScheme,
			zcncore.WithMinSubmit(chain.MinSubmit),
			zcncore.WithMinConfirmation(chain.MinConfirmation),
			zcncore.WithConfirmationChainLength(chain.ConfirmationChainLength),
			zcncore.WithChainID(chain.ChainID),
			zcncore.WithEthereumNode(chain.EthNode))

	}
	zcncore.Logger.Info("******* Wallet SDK Version:", version.VERSIONSTR, " ******* (Init)")
	return err
}

// GetLockConfig returns the lock token configuration information such as interest rate from blockchain
func GetLockConfig(cb GetInfoCallback) error {
	err := zcncore.CheckConfig()
	if err != nil {
		return err
	}
	go zcncore.GetInfoFromSharders(GET_LOCK_CONFIG, OpGetTokenLockConfig, cb)
	return nil
}

// GetLockedTokens returns the ealier locked token pool stats
func GetLockedTokens(cb GetInfoCallback) error {
	err := zcncore.CheckConfig()
	if err != nil {
		return err
	}
	go func() {
		urlSuffix := fmt.Sprintf("%v%v", GET_LOCKED_TOKENS, zcncore.GetClientWalletID())
		zcncore.GetInfoFromSharders(urlSuffix, OpGetLockedTokens, cb)
	}()
	return nil
}

type Params map[string]string

func (p Params) Query() string {
	if len(p) == 0 {
		return ""
	}
	var params = make(url.Values)
	for k, v := range p {
		params[k] = []string{v}
	}
	return "?" + params.Encode()
}

func withParams(uri string, params Params) string {
	return uri + params.Query()
}

// interest pools sc

func GetInterestPoolSCConfig(cb GetInfoCallback) (err error) {
	if err = zcncore.CheckConfig(); err != nil {
		return
	}
	go zcncore.GetInfoFromSharders(GET_INTERESTPOOLSC_CONFIG, 0, cb)
	return
}

// GetWritePoolInfo obtains information about all write pools of a user.
// If given clientID is empty, then current user used.
func GetWritePoolInfo(clientID string, cb GetInfoCallback) (err error) {

	if err = zcncore.CheckConfig(); err != nil {
		return
	}
	if clientID == "" {
		clientID = zcncore.GetClientWalletID()
	}

	var url = withParams(STORAGESC_GET_WRITE_POOL_INFO, Params{
		"client_id": clientID,
	})
	go zcncore.GetInfoFromSharders(url, OpStorageSCGetWritePoolInfo, cb)
	return
}
