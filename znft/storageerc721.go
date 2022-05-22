package znft

import (
	"context"
	"math/big"

	dstorageerc721 "github.com/0chain/gosdk/znft/contracts/dstorageerc721/binding"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	ContractStorageERC721Name = "StorageERC721"
	Withdraw                  = "withdraw"
	SetReceiver               = "setReceiver"
	SetRoyalty                = "setRoyalty"
	SetMintable               = "setMintable"
	SetMax                    = "setMax"
	SetAllocation             = "setAllocation"
	SetURI                    = "setURI"
	TokenURIFallback          = "tokenURIFallback"
	Price                     = "price"
	Mint                      = "mint"
	MintOwner                 = "mintOwner"
	RoyaltyInfo               = "royaltyInfo"
)

// Solidity Functions
// - withdraw()
// - setReceiver(address receiver_)
// - setRoyalty(uint256 royalty_)
// - setMintable(bool status_)
// - setMax(uint256 max_)
// - setAllocation(string calldata allocation_)
// - setURI(string calldata uri_)
// - tokenURIFallback(uint256 tokenId)  returns (string memory)
// - price() returns (uint256)
// - mint(uint256 amount)
// - mintOwner(uint256 amount)
// - royaltyInfo(uint256 tokenId, uint256 salePrice) returns (address, uint256)

type IStorageECR721 interface {
	Withdraw() error
	SetReceiver(receiver string) error
	SetRoyalty(sum *big.Int) error
	SetMintable(status bool) error
	SetMax(max *big.Int) error
	SetAllocation(allocation string) error
	SetURI(uri string) error
	TokenURIFallback(token *big.Int) (string, error)
	Price() (int64, error)
	Mint(amount *big.Int) error
	MintOwner(amount *big.Int) error
	RoyaltyInfo(tokenId, salePrice *big.Int) (string, *big.Int, error)
}

type StorageECR721 struct {
	session *dstorageerc721.BindingsSession
	ctx     context.Context
}

var (
	_ IStorageECR721 = (*StorageECR721)(nil)
)

func (conf *StorageECR721) Mint(amount *big.Int) error {
	evmTr, err := conf.session.Mint(amount)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", Mint)
	}

	evmTr.Hash()

	return nil
}

func (conf *StorageECR721) RoyaltyInfo(tokenId, salePrice *big.Int) (string, *big.Int, error) {
	address, sum, err := conf.session.RoyaltyInfo(tokenId, salePrice)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to execute %s", RoyaltyInfo)
	}

	return address.Hex(), sum, nil
}

func (conf *StorageECR721) MintOwner(amount *big.Int) error {
	evmTr, err := conf.session.MintOwner(amount)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", MintOwner)
	}

	evmTr.Hash()

	return nil
}

func (conf *StorageECR721) TokenURIFallback(token *big.Int) (string, error) {
	tokenURI, err := conf.session.TokenURIFallback(token)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute %s", TokenURIFallback)
	}

	return tokenURI, nil
}

// Price returns price
func (conf *StorageECR721) Price() (int64, error) {
	price, err := conf.session.Price()
	if err != nil {
		return -1, errors.Wrapf(err, "failed to execute %s", Price)
	}

	return price.Int64(), nil
}

// SetURI updates uri
func (conf *StorageECR721) SetURI(uri string) error {
	evmTr, err := conf.session.SetURI(uri)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetURI)
	}

	evmTr.Hash()

	return nil
}

// SetAllocation updates allocation
func (conf *StorageECR721) SetAllocation(allocation string) error {
	evmTr, err := conf.session.SetAllocation(allocation)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetAllocation)
	}

	evmTr.Hash()

	return nil
}

// SetMax eth balance from token contract - setReceiver(address receiver_)
func (conf *StorageECR721) SetMax(max *big.Int) error {
	evmTr, err := conf.session.SetMax(max)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetMax)
	}

	evmTr.Hash()

	return nil
}

// SetMintable updates mintable state
func (conf *StorageECR721) SetMintable(status bool) error {
	evmTr, err := conf.session.SetMintable(status)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetMintable)
	}

	evmTr.Hash()

	return nil
}

// SetRoyalty eth balance from token contract - setReceiver(address receiver_)
func (conf *StorageECR721) SetRoyalty(sum *big.Int) error {
	evmTr, err := conf.session.SetRoyalty(sum)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetRoyalty)
	}

	evmTr.Hash()

	return nil
}

// SetReceiver eth balance from token contract - setReceiver(address receiver_)
func (conf *StorageECR721) SetReceiver(receiver string) error {
	address := common.HexToAddress(receiver)

	evmTr, err := conf.session.SetReceiver(address)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetReceiver)
	}

	evmTr.Hash()

	return nil
}

// Withdraw eth balance from token contract - withdraw()
func (conf *StorageECR721) Withdraw() error {
	evmTr, err := conf.session.Withdraw()
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", Withdraw)
	}

	evmTr.Hash()

	return nil
}
