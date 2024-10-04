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
func MinerStakePoolLock(providerId string, lock int64) ([]byte, error) {
	if tokenValidation(lock) != nil {
		return nil, tokenValidation(lock)
	}

	_, _, _, txn, err := zcncore.MinerSCLock(providerId, zcncore.ProviderMiner, uint64(lock))

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
func MinerStakePoolUnlock(providerId string) ([]byte, error) {
	_, _, _, txn, err := zcncore.MinerSCUnlock(providerId, zcncore.ProviderMiner)

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
func MinerCollectReward(providerId string) ([]byte, error) {
	_, _, _, txn, err := zcncore.MinerSCCollectReward(providerId, zcncore.ProviderMiner)

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
func BlobberStakePoolLock(providerId string, lock int64) ([]byte, error) {
	if tokenValidation(lock) != nil {
		return nil, tokenValidation(lock)
	}

	_, _, _, txn, err := zcncore.StorageSCStakePoolLock(providerId, zcncore.ProviderBlobber, uint64(lock))

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
func BlobberStakePoolUnlock(providerId string) ([]byte, error) {
	_, _, _, txn, err := zcncore.StorageSCStakePoolUnlock(providerId, zcncore.ProviderBlobber)

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
func BlobberCollectReward(providerId string) ([]byte, error) {
	_, _, _, txn, err := zcncore.StorageSCCollectReward(providerId, zcncore.ProviderBlobber)

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
func SharderStakePoolLock(providerId string, lock int64) ([]byte, error) {
	if tokenValidation(lock) != nil {
		return nil, tokenValidation(lock)
	}

	_, _, _, txn, err := zcncore.MinerSCLock(providerId, zcncore.ProviderSharder, uint64(lock))

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
func SharderStakePoolUnlock(providerId string) ([]byte, error) {
	_, _, _, txn, err := zcncore.MinerSCUnlock(providerId, zcncore.ProviderSharder)

	if err != nil {
		return nil, errors.New("Error in SharderSCUnlock: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

func SharderCollectReward(providerId string) ([]byte, error) {
	_, _, _, txn, err := zcncore.MinerSCCollectReward(providerId, zcncore.ProviderSharder)

	if err != nil {
		return nil, errors.New("Error in SharderSCCollectReward: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil
}

func SendZCN(senderClientId string, tokens int64, description string) ([]byte, error) {
	if tokenValidation(tokens) != nil {
		return nil, tokenValidation(tokens)

	}

	_, _, _, txn, err := zcncore.Send(senderClientId, uint64(tokens), description)

	if err != nil {
		return nil, errors.New("Error in SendZCN: " + err.Error())
	}
	byteTxn, err := json.Marshal(txn)
	if err != nil {
		return nil, errors.New("Error in Encoding txn: " + err.Error())
	}
	return byteTxn, nil

}

// Faucet sends tokens to the user.
// # Inputs
//   - tokens: tokens to send
//   - input: input
func Faucet(tokens int64, input string) ([]byte, error) {
	if tokenValidation(tokens) != nil {
		return nil, tokenValidation(tokens)
	}

	_, _, _, txn, err := zcncore.Faucet(uint64(tokens), input)

	if err != nil {
		return nil, errors.New("Error in Faucet: " + err.Error())
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
