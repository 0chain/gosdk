package zcn

import (
	"encoding/json"
	"errors"
	"github.com/0chain/gosdk/zcncore"
)

// MinerStakePoolLock locks the stake pool for the miner.
// # Inputs
//   - providerId: provider id
//   - lock: lock amount
func MinerStakePoolLock(providerId string, lock int64, fee int64) ([]byte, error) {
	if tokenValidation(lock) != nil {
		return nil, tokenValidation(lock)
	}

	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.MinerSCLock(providerId, zcncore.ProviderMiner, uint64(lock), uint64(fee))

	if err != nil {
		return nil, errors.New("Error in MinerSCLock: " + err.Error())
	}

	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

// MinerStakePoolUnlock unlocks the stake pool for the miner.
// # Inputs
//   - providerId: provider id
func MinerStakePoolUnlock(providerId string, fee int64) ([]byte, error) {
	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.MinerSCUnlock(providerId, zcncore.ProviderMiner, uint64(fee))

	if err != nil {
		return nil, errors.New("Error in MinerSCUnlock: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

// MinerCollectReward collects the reward from the miner smart contract.
// # Inputs
//   - providerId: provider id
func MinerCollectReward(providerId string, fee int64) ([]byte, error) {
	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.MinerSCCollectReward(providerId, zcncore.ProviderMiner, uint64(fee))

	if err != nil {
		return nil, errors.New("Error in MinerSCCollectReward: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

// BlobberStakePoolLock locks the stake pool for the blobber.
// # Inputs
//   - providerId: provider id
//   - lock: lock amount
func BlobberStakePoolLock(providerId string, lock int64, fee int64) ([]byte, error) {
	if tokenValidation(lock) != nil {
		return nil, tokenValidation(lock)
	}

	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.StorageSCStakePoolLock(providerId, zcncore.ProviderBlobber, uint64(lock), uint64(fee))

	if err != nil {
		return nil, errors.New("Error in BlobberSCLock: " + err.Error())
	}

	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

// BlobberStakePoolUnlock unlocks the stake pool for the blobber.
// # Inputs
//   - providerId: provider id
func BlobberStakePoolUnlock(providerId string, fee int64) ([]byte, error) {
	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.StorageSCStakePoolUnlock(providerId, zcncore.ProviderBlobber, uint64(fee))

	if err != nil {
		return nil, errors.New("Error in BlobberSCUnlock: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

// BlobberCollectReward locks the stake pool for the blobber.
// # Inputs
//   - providerId: provider id
func BlobberCollectReward(providerId string, fee int64) ([]byte, error) {
	_, _, _, txn, err := zcncore.StorageSCCollectReward(providerId, zcncore.ProviderBlobber, uint64(fee))

	if err != nil {
		return nil, errors.New("Error in BlobberSCCollectReward: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

// SharderStakePoolLock locks the stake pool for the sharder.
// # Inputs
//   - providerId: provider id
//   - lock: lock amount
func SharderStakePoolLock(providerId string, lock int64, fee int64) ([]byte, error) {
	if tokenValidation(lock) != nil {
		return nil, tokenValidation(lock)
	}

	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.MinerSCLock(providerId, zcncore.ProviderSharder, uint64(lock), uint64(fee))

	if err != nil {
		return nil, errors.New("Error in SharderSCLock: " + err.Error())
	}

	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

// SharderStakePoolUnlock unlocks the stake pool for the sharder.
// # Inputs
//   - providerId: provider id
func SharderStakePoolUnlock(providerId string, fee int64) ([]byte, error) {
	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.MinerSCUnlock(providerId, zcncore.ProviderSharder, uint64(fee))

	if err != nil {
		return nil, errors.New("Error in SharderSCUnlock: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

func SharderCollectReward(providerId string, fee int64) ([]byte, error) {
	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.MinerSCCollectReward(providerId, zcncore.ProviderSharder, uint64(fee))

	if err != nil {
		return nil, errors.New("Error in SharderSCCollectReward: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

func SendZCN(senderClientId string, tokens int64, fee int64, description string) ([]byte, error) {
	if tokenValidation(tokens) != nil {
		return nil, tokenValidation(tokens)

	}

	if tokenValidation(fee) != nil {
		return nil, tokenValidation(fee)
	}

	_, _, _, txn, err := zcncore.Send(senderClientId, uint64(tokens), uint64(fee), description)

	if err != nil {
		return nil, errors.New("Error in SendZCN: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil

}

// VerifyTransaction verifies the transaction.
// # Inputs
//   - hash: hash of the transaction
func VerifyTransaction(hash string) ([]byte, error) {
	txn, err := zcncore.VerifyTransaction(hash)
	if err != nil {
		return nil, errors.New("Error in VerifyTransaction: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

func tokenValidation(tokens int64) error {
	if tokens <= 0 {
		return errors.New("negative or zero tokens not allowed")
	}
	return nil
}

func getProviderType(providerType int) zcncore.Provider {
	switch providerType {
	case 0:
		return zcncore.ProviderMiner
	case 1:
		return zcncore.ProviderSharder
	}
	return zcncore.ProviderMiner
}
