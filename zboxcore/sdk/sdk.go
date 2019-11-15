package sdk

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"time"

	"github.com/0chain/gosdk/zboxcore/marker"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const STORAGE_SCADDRESS = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7"

const (
	OpUpload   int = 0
	OpDownload int = 1
	OpRepair   int = 2
)

type StatusCallback interface {
	Started(allocationId, filePath string, op int, totalBytes int)
	InProgress(allocationId, filePath string, op int, completedBytes int)
	Error(allocationID string, filePath string, op int, err error)
	Completed(allocationId, filePath string, filename string, mimetype string, size int, op int)
}

var numBlockDownloads = 10
var sdkInitialized = false

// GetVersion - returns version string
func GetVersion() string {
	return version.VERSIONSTR
}

// logFile - Log file
// verbose - true - console output; false - no console output
func SetLogFile(logFile string, verbose bool) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	Logger.SetLogFile(f, verbose)
	Logger.Info("******* Storage SDK Version: ", version.VERSIONSTR, " *******")
}

func InitStorageSDK(clientJson string, miners []string, sharders []string, chainID string, signatureScheme string) error {
	err := client.PopulateClient(clientJson, signatureScheme)
	if err != nil {
		return err
	}
	blockchain.SetMiners(miners)
	blockchain.SetSharders(sharders)
	blockchain.SetChainID(chainID)
	sdkInitialized = true
	return nil
}

func GetClientEncryptedPublicKey() (string, error) {
	if !sdkInitialized {
		return "", common.NewError("sdk_not_initialized", "SDK is not initialised")
	}
	encScheme := encryption.NewEncryptionScheme()
	err := encScheme.Initialize(client.GetClient().Mnemonic)
	if err != nil {
		return "", err
	}
	return encScheme.GetPublicKey()
}

func GetAllocationFromAuthTicket(authTicket string) (*Allocation, error) {
	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, common.NewError("auth_ticket_decode_error", "Error unmarshaling the auth ticket."+err.Error())
	}
	return GetAllocation(at.AllocationID)
}

func GetAllocation(allocationID string) (*Allocation, error) {
	params := make(map[string]string)
	params["allocation"] = allocationID
	allocationBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocation", params, nil)
	if err != nil {
		return nil, common.NewError("allocation_fetch_error", "Error fetching the allocation."+err.Error())
	}
	allocationObj := &Allocation{}
	err = json.Unmarshal(allocationBytes, allocationObj)
	if err != nil {
		return nil, common.NewError("allocation_decode_error", "Error decoding the allocation."+err.Error())
	}
	allocationObj.numBlockDownloads = numBlockDownloads
	allocationObj.InitAllocation()
	return allocationObj, nil
}

func SetNumBlockDownloads(num int) {
	if num > 0 && num <= 100 {
		numBlockDownloads = num
	}
	return
}

func GetAllocations() ([]*Allocation, error) {
	return GetAllocationsForClient(client.GetClientID())
}

func GetAllocationsForClient(clientID string) ([]*Allocation, error) {
	params := make(map[string]string)
	params["client"] = clientID
	allocationsBytes, err := zboxutil.MakeSCRestAPICall(STORAGE_SCADDRESS, "/allocations", params, nil)
	if err != nil {
		return nil, common.NewError("allocations_fetch_error", "Error fetching the allocations."+err.Error())
	}
	allocations := make([]*Allocation, 0)
	err = json.Unmarshal(allocationsBytes, &allocations)
	if err != nil {
		return nil, common.NewError("allocations_decode_error", "Error decoding the allocations."+err.Error())
	}
	return allocations, nil
}

func CreateAllocation(datashards int, parityshards int, size int64, expiry int64) (string, error) {
	return CreateAllocationForOwner(client.GetClientID(), client.GetClientPublicKey(), datashards, parityshards, size, expiry)
}

func CreateAllocationForOwner(owner string, ownerpublickey string, datashards int, parityshards int, size int64, expiry int64) (string, error) {
	allocationRequest := make(map[string]interface{})
	allocationRequest["data_shards"] = datashards
	allocationRequest["parity_shards"] = parityshards
	allocationRequest["size"] = size
	allocationRequest["owner_id"] = owner
	allocationRequest["owner_public_key"] = ownerpublickey
	allocationRequest["expiration_date"] = expiry

	sn := transaction.SmartContractTxnData{Name: transaction.NEW_ALLOCATION_REQUEST, InputArgs: allocationRequest}
	allocationRequestBytes, err := json.Marshal(sn)
	if err != nil {
		return "", err
	}
	txn := transaction.NewTransactionEntity(client.GetClientID(), blockchain.GetChainID(), client.GetClientPublicKey())
	txn.TransactionData = string(allocationRequestBytes)
	txn.ToClientID = STORAGE_SCADDRESS
	txn.Value = 0
	txn.TransactionType = transaction.TxnTypeSmartContract
	err = txn.ComputeHashAndSign(client.Sign)
	if err != nil {
		return "", err
	}
	transaction.SendTransactionSync(txn, blockchain.GetMiners())
	time.Sleep(5 * time.Second)
	retries := 0
	var t *transaction.Transaction
	for retries < 5 {
		t, err = transaction.VerifyTransaction(txn.Hash, blockchain.GetSharders())
		if err == nil {
			break
		}
		retries++
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		Logger.Error("Error verifying the allocation transaction", err.Error(), txn.Hash)
		return "", err
	}
	if t == nil {
		return "", common.NewError("transaction_validation_failed", "Failed to get the transaction confirmation")
	}

	return t.Hash, nil
}
