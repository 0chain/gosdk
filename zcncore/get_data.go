package zcncore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/0chain/gosdk/core/block"
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/tokenrate"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"net/url"
	"strconv"
)

type GetClientResponse struct {
	ID           string `json:"id"`
	Version      string `json:"version"`
	CreationDate int    `json:"creation_date"`
	PublicKey    string `json:"public_key"`
}

func GetClientDetails(clientID string) (*GetClientResponse, error) {
	clientNode, err := client.GetNode()
	if err != nil {
		panic(err)
	}
	minerurl := util.GetRandom(clientNode.Network().Miners, 1)[0]
	url := minerurl + GET_CLIENT
	url = fmt.Sprintf("%v?id=%v", url, clientID)
	req, err := util.NewHTTPGetRequest(url)
	if err != nil {
		logging.Error(minerurl, "new get request failed. ", err.Error())
		return nil, err
	}
	res, err := req.Get()
	if err != nil {
		logging.Error(minerurl, "send error. ", err.Error())
		return nil, err
	}

	var clientDetails GetClientResponse
	err = json.Unmarshal([]byte(res.Body), &clientDetails)
	if err != nil {
		return nil, err
	}

	return &clientDetails, nil
}

// Deprecated: Use zcncrypto.IsMnemonicValid()
// IsMnemonicValid is an utility function to check the mnemonic valid
//
//	# Inputs
//	-	mnemonic: mnemonics
func IsMnemonicValid(mnemonic string) bool {
	return zcncrypto.IsMnemonicValid(mnemonic)
}

// SetWalletInfo should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
//
//	# Inputs
//	- jsonWallet: json format of wallet
//	{
//	"client_id":"30764bcba73216b67c36b05a17b4dd076bfdc5bb0ed84856f27622188c377269",
//	"client_key":"1f495df9605a4479a7dd6e5c7a78caf9f9d54e3a40f62a3dd68ed377115fe614d8acf0c238025f67a85163b9fbf31d10fbbb4a551d1cf00119897edf18b1841c",
//	"keys":[
//		{"public_key":"1f495df9605a4479a7dd6e5c7a78caf9f9d54e3a40f62a3dd68ed377115fe614d8acf0c238025f67a85163b9fbf31d10fbbb4a551d1cf00119897edf18b1841c","private_key":"41729ed8d82f782646d2d30b9719acfd236842b9b6e47fee12b7bdbd05b35122"}
//	],
//	"mnemonics":"glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp",
//	"version":"1.0",
//	"date_created":"1662534022",
//	"nonce":0
//	}
//
// - splitKeyWallet: if wallet keys is split
func SetWalletInfo(jsonWallet, sigScheme string, splitKeyWallet bool) error {
	wallet := zcncrypto.Wallet{}
	err := json.Unmarshal([]byte(jsonWallet), &wallet)
	if err != nil {
		return errors.New("invalid jsonWallet: " + err.Error())
	}

	client.SetWallet(wallet)
	client.SetSignatureScheme(sigScheme)

	return client.SetSplitKeyWallet(splitKeyWallet)
}

// SetAuthUrl will be called by app to set zauth URL to SDK.
// # Inputs
//   - url: the url of zAuth server
func SetAuthUrl(url string) error {
	return client.SetAuthUrl(url)
}

// ConvertTokenToUSD converts the ZCN tokens to USD amount
//   - token: ZCN tokens amount
func ConvertTokenToUSD(token float64) (float64, error) {
	zcnRate, err := getTokenUSDRate()
	if err != nil {
		return 0, err
	}
	return token * zcnRate, nil
}

// ConvertUSDToToken converts the USD amount to ZCN tokens
//   - usd: USD amount
func ConvertUSDToToken(usd float64) (float64, error) {
	zcnRate, err := getTokenUSDRate()
	if err != nil {
		return 0, err
	}
	return usd * (1 / zcnRate), nil
}

func getTokenUSDRate() (float64, error) {
	return tokenrate.GetUSD(context.TODO(), "zcn")
}

// getWallet get a wallet object from a wallet string
func getWallet(walletStr string) (*zcncrypto.Wallet, error) {
	var w zcncrypto.Wallet
	err := json.Unmarshal([]byte(walletStr), &w)
	if err != nil {
		fmt.Printf("error while parsing wallet string.\n%v\n", err)
		return nil, err
	}

	return &w, nil
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

func withParams(uri string, params Params) string { //nolint:unused
	return uri + params.Query()
}

// GetBlobberSnapshots obtains list of allocations of a blobber.
// Blobber snapshots are historical records of the blobber instance to track its change over time and serve graph requests,
// which are requests that need multiple data points, distributed over an interval of time, usually to plot them on a
// graph.
//   - round: round number
//   - limit: how many blobber snapshots should be fetched
//   - offset: how many blobber snapshots should be skipped
//   - cb: info callback instance, carries the response of the GET request to the sharders
//func GetBlobberSnapshots(round int64, limit int64, offset int64) (res []byte, err error) {
//	if err = CheckConfig(); err != nil {
//		return
//	}
//
//	return coreHttp.MakeSCRestAPICall(StorageSmartContractAddress, STORAGE_GET_BLOBBER_SNAPSHOT, Params{
//		"round":  strconv.FormatInt(round, 10),
//		"limit":  strconv.FormatInt(limit, 10),
//		"offset": strconv.FormatInt(offset, 10),
//	}, nil)
//}

// GetMinerSCNodeInfo get miner information from sharders
//   - id: the id of miner
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMinerSCNodeInfo(id string) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}

	return client.MakeSCRestAPICall(MinerSmartContractAddress, GET_MINERSC_NODE, Params{
		"id": id,
	})
}

// GetMintNonce retrieve the client's latest mint nonce from sharders
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetMintNonce() ([]byte, error) {
	err := CheckConfig()
	if err != nil {
		return nil, err
	}

	return client.MakeSCRestAPICall(ZCNSCSmartContractAddress, GET_MINT_NONCE, Params{
		"client_id": client.ClientID(),
	})
}

func GetMiners(active, stakable bool, limit, offset int) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}

	return client.MakeSCRestAPICall(MinerSmartContractAddress, GET_MINERSC_MINERS, Params{
		"active":   strconv.FormatBool(active),
		"stakable": strconv.FormatBool(stakable),
		"offset":   strconv.FormatInt(int64(offset), 10),
		"limit":    strconv.FormatInt(int64(limit), 10),
	})
}

func GetSharders(active, stakable bool, limit, offset int) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}

	return client.MakeSCRestAPICall(MinerSmartContractAddress, GET_MINERSC_SHARDERS, Params{
		"active":   strconv.FormatBool(active),
		"stakable": strconv.FormatBool(stakable),
		"offset":   strconv.FormatInt(int64(offset), 10),
		"limit":    strconv.FormatInt(int64(limit), 10),
	})
}

func GetBlobbers(active, stakable bool, limit, offset int) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}

	blobbers, err := sdk.GetBlobbersPaged(active, stakable, limit, offset)

	if err != nil {
		return nil, err
	}

	blobbersBytes, err := json.Marshal(blobbers)
	if err != nil {
		return nil, err
	}

	return blobbersBytes, nil
}

// GetLatestFinalizedMagicBlock gets latest finalized magic block
//   - numSharders: number of sharders
//   - timeout: request timeout
func GetLatestFinalizedMagicBlock() (m *block.MagicBlock, err error) {
	res, err := client.MakeSCRestAPICall("", GET_LATEST_FINALIZED_MAGIC_BLOCK, nil, "")
	if err != nil {
		return nil, err
	}

	type respObj struct {
		MagicBlock *block.MagicBlock `json:"magic_block"`
	}

	var resp respObj

	err = json.Unmarshal(res, &resp)
	if err != nil {
		return nil, err
	}

	return resp.MagicBlock, nil
}

// GetMinerSCUserInfo retrieve user stake pools for the providers related to the Miner SC (miners/sharders).
//   - clientID: user's wallet id
func GetMinerSCUserInfo(clientID string) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}
	if clientID == "" {
		clientID = client.ClientID()
	}

	return client.MakeSCRestAPICall(MinerSmartContractAddress, GET_MINERSC_USER, Params{
		"client_id": clientID,
	})
}

// GetMinerSCNodePool get miner smart contract node pool
//   - id: the id of miner
func GetMinerSCNodePool(id string) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}

	return client.MakeSCRestAPICall(MinerSmartContractAddress, GET_MINERSC_POOL, Params{
		"id":      id,
		"pool_id": client.ClientID(),
	})
}

// GetNotProcessedZCNBurnTickets retrieve burn tickets that are not compensated by minting
//   - ethereumAddress: ethereum address for the issuer of the burn tickets
//   - startNonce: start nonce for the burn tickets
//   - cb: info callback instance, carries the response of the GET request to the sharders
func GetNotProcessedZCNBurnTickets(ethereumAddress, startNonce string) ([]byte, error) {
	err := CheckConfig()
	if err != nil {
		return nil, err
	}

	const GET_NOT_PROCESSED_BURN_TICKETS = `/v1/not_processed_burn_tickets`

	return client.MakeSCRestAPICall(ZCNSCSmartContractAddress, GET_NOT_PROCESSED_BURN_TICKETS, Params{
		"ethereum_address": ethereumAddress,
		"nonce":            startNonce,
	})
}

// GetUserLockedTotal get total token user locked
// # Inputs
//   - clientID wallet id
func GetUserLockedTotal(clientID string) (int64, error) {
	var result map[string]int64

	err := CheckConfig()
	if err != nil {
		return 0, err
	}

	const GET_USER_LOCKED_TOTAL = `/v1/getUserLockedTotal`

	info, err := client.MakeSCRestAPICall(ZCNSCSmartContractAddress, GET_USER_LOCKED_TOTAL, Params{
		"client_id": clientID,
	})

	if err != nil {
		return 0, errors.New("error while making rest api call: " + err.Error())
	}

	err = json.Unmarshal([]byte(info), &result)
	if err != nil {
		return 0, errors.New("invalid json format: " + err.Error())
	}

	total, ok := result["total"]
	if ok {
		return total, nil
	} else {
		return 0, err
	}
}

// GetStakePoolUserInfo get stake pool user info for all blobbers/miners/sharders.
// # Inputs
//   - clientID wallet id
func GetStakePoolUserInfo(clientID string) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}

	limit, offset := 20, 0
	spUserInfo, err := sdk.GetStakePoolUserInfo(clientID, limit, offset)

	if err != nil {
		return nil, err
	}

	var spUserInfoSl []*sdk.StakePoolUserInfo
	spUserInfoSl = append(spUserInfoSl, spUserInfo)
	for {
		// if the length of the slice is less than the limit, then we have reached the end
		if len(spUserInfoSl) < limit {
			break
		}

		// get the next set of stake pool user info
		offset += limit
		spUserInfo, err = sdk.GetStakePoolUserInfo(clientID, limit, offset)
		if err != nil {
			break
		}
		spUserInfoSl = append(spUserInfoSl, spUserInfo)
	}

	spUserInfoBytes, err := json.Marshal(spUserInfoSl)
	if err != nil {
		return nil, errors.New("error while marshalling stake pool user info: " + err.Error())
	}

	return spUserInfoBytes, nil
}

func GetTransactions(toClientId, fromClientId, order string, limit, offset int64) ([]byte, error) {
	if err := CheckConfig(); err != nil {
		return nil, err
	}

	const GET_TRANSACTIONS = `/transactions`

	return client.MakeSCRestAPICall(StorageSmartContractAddress, GET_TRANSACTIONS, Params{
		"to_client_id": toClientId,
		"client_id":    fromClientId,
		"order":        order,
		"limit":        strconv.FormatInt(limit, 10),
		"offset":       strconv.FormatInt(offset, 10),
	})
}

func GetFeesTable(reqPercentage float32) ([]byte, error) {
	nodeClient, err := client.GetNode()
	if err != nil {
		return nil, err
	}
	fees, err := transaction.GetFeesTable(nodeClient.GetStableMiners(), reqPercentage)

	if err != nil {
		return nil, err
	}

	feesBytes, err := json.Marshal(fees)
	if err != nil {
		return nil, err
	}

	return feesBytes, nil
}
