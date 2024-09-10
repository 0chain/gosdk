package transaction

import (
	"encoding/json"
	"github.com/0chain/errors"
	coreHttp "github.com/0chain/gosdk/core/client"
)

const (
	StorageSmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`
	FaucetSmartContractAddress  = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d3`
	MinerSmartContractAddress   = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9`
	ZCNSCSmartContractAddress   = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0`
)
const (
	GET_CLIENT                       = `/v1/client/get`
	PUT_TRANSACTION                  = `/v1/transaction/put`
	GET_BLOCK_INFO                   = `/v1/block/get?`
	GET_MAGIC_BLOCK_INFO             = `/v1/block/magic/get?`
	GET_LATEST_FINALIZED             = `/v1/block/get/latest_finalized`
	GET_LATEST_FINALIZED_MAGIC_BLOCK = `/v1/block/get/latest_finalized_magic_block`
	GET_FEE_STATS                    = `/v1/block/get/fee_stats`
	GET_CHAIN_STATS                  = `/v1/chain/get/stats`

	// zcn sc
	ZCNSC_PFX                      = `/v1/screst/` + ZCNSCSmartContractAddress
	GET_MINT_NONCE                 = ZCNSC_PFX + `/v1/mint_nonce`
	GET_NOT_PROCESSED_BURN_TICKETS = ZCNSC_PFX + `/v1/not_processed_burn_tickets`
	GET_AUTHORIZER                 = ZCNSC_PFX + `/getAuthorizer`

	// miner SC

	MINERSC_PFX          = `/v1/screst/` + MinerSmartContractAddress
	GET_MINERSC_NODE     = MINERSC_PFX + "/nodeStat"
	GET_MINERSC_POOL     = MINERSC_PFX + "/nodePoolStat"
	GET_MINERSC_CONFIG   = MINERSC_PFX + "/configs"
	GET_MINERSC_GLOBALS  = MINERSC_PFX + "/globalSettings"
	GET_MINERSC_USER     = MINERSC_PFX + "/getUserPools"
	GET_MINERSC_MINERS   = MINERSC_PFX + "/getMinerList"
	GET_MINERSC_SHARDERS = MINERSC_PFX + "/getSharderList"
	GET_MINERSC_EVENTS   = MINERSC_PFX + "/getEvents"

	// storage SC

	STORAGESC_PFX = "/v1/screst/" + StorageSmartContractAddress

	STORAGESC_GET_SC_CONFIG            = STORAGESC_PFX + "/storage-config"
	STORAGESC_GET_CHALLENGE_POOL_INFO  = STORAGESC_PFX + "/getChallengePoolStat"
	STORAGESC_GET_ALLOCATION           = STORAGESC_PFX + "/allocation"
	STORAGESC_GET_ALLOCATIONS          = STORAGESC_PFX + "/allocations"
	STORAGESC_GET_READ_POOL_INFO       = STORAGESC_PFX + "/getReadPoolStat"
	STORAGESC_GET_STAKE_POOL_INFO      = STORAGESC_PFX + "/getStakePoolStat"
	STORAGESC_GET_STAKE_POOL_USER_INFO = STORAGESC_PFX + "/getUserStakePoolStat"
	STORAGESC_GET_USER_LOCKED_TOTAL    = STORAGESC_PFX + "/getUserLockedTotal"
	STORAGESC_GET_BLOBBERS             = STORAGESC_PFX + "/getblobbers"
	STORAGESC_GET_BLOBBER              = STORAGESC_PFX + "/getBlobber"
	STORAGESC_GET_VALIDATOR            = STORAGESC_PFX + "/get_validator"
	STORAGESC_GET_TRANSACTIONS         = STORAGESC_PFX + "/transactions"

	STORAGE_GET_SNAPSHOT            = STORAGESC_PFX + "/replicate-snapshots"
	STORAGE_GET_BLOBBER_SNAPSHOT    = STORAGESC_PFX + "/replicate-blobber-aggregates"
	STORAGE_GET_MINER_SNAPSHOT      = STORAGESC_PFX + "/replicate-miner-aggregates"
	STORAGE_GET_SHARDER_SNAPSHOT    = STORAGESC_PFX + "/replicate-sharder-aggregates"
	STORAGE_GET_AUTHORIZER_SNAPSHOT = STORAGESC_PFX + "/replicate-authorizer-aggregates"
	STORAGE_GET_VALIDATOR_SNAPSHOT  = STORAGESC_PFX + "/replicate-validator-aggregates"
	STORAGE_GET_USER_SNAPSHOT       = STORAGESC_PFX + "/replicate-user-aggregates"
)

//
// storage SC configurations and blobbers
//

type InputMap struct {
	Fields map[string]string `json:"fields"`
}

// GetStorageSCConfig retrieves storage SC configurations.
func GetConfig(configType string) (conf *InputMap, err error) {
	var (
		scAddress    string
		relativePath string
		b            []byte
	)

	if configType == "storage_sc_config" {
		scAddress = StorageSmartContractAddress
		relativePath = STORAGESC_GET_SC_CONFIG
	} else if configType == "miner_sc_globals" {
		scAddress = MinerSmartContractAddress
		relativePath = GET_MINERSC_GLOBALS
	}

	b, err = coreHttp.MakeSCRestAPICall(scAddress, relativePath, nil,
		nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting storage SC configs:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	conf = new(InputMap)
	conf.Fields = make(map[string]string)
	if err = json.Unmarshal(b, conf); err != nil {
		return nil, errors.Wrap(err, "error decoding response:")
	}

	return
}
