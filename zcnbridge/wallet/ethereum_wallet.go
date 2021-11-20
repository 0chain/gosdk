package wallet

import (
	"crypto/ecdsa"

	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type EthereumWallet struct {
	PublicKey  *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

func NewEthereumWallet() (*EthereumWallet, error) {
	address, publicKey, privateKey, err := ethereum.GetKeysAddress()
	if err != nil {
		return nil, errors.Wrap("wallet_init", "failed to initialize ethereum wallet", err)
	}

	return &EthereumWallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}
