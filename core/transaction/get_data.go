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
	GET_MINERSC_GLOBALS     = "/configs"
	STORAGESC_GET_SC_CONFIG = "/storage-config"
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

	b, err = coreHttp.MakeSCRestAPICall(scAddress, relativePath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting storage SC configs:")
	}
	if len(b) == 0 {
		return nil, errors.New("", "empty response")
	}

	conf = new(InputMap)
	conf.Fields = make(map[string]string)
	if err = json.Unmarshal(b, conf); err != nil {
		return nil, errors.Wrap(err, "1 error decoding response:")
	}

	return
}
