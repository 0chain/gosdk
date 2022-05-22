package znft

import (
	"context"
	"math/big"

	storageerc721 "github.com/0chain/gosdk/znft/contracts/dstorageerc721/binding"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	ContractStorageERC721Name       = "StorageERC721"
	ContractStorageERC721FixedName  = "StorageERC721Fixed"
	ContractStorageERC721PackName   = "StorageERC721Pack"
	ContractStorageERC721RandomName = "StorageERC721Random"
	Withdraw                        = "withdraw"
	SetReceiver                     = "setReceiver"
	SetRoyalty                      = "setRoyalty"
	SetMintable                     = "setMintable"
	SetMax                          = "setMax"
	SetAllocation                   = "setAllocation"
	SetURI                          = "setURI"
	TokenURIFallback                = "tokenURIFallback"
	Price                           = "price"
	Mint                            = "mint"
	MintOwner                       = "mintOwner"
	RoyaltyInfo                     = "royaltyInfo"
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
// Fields:
//    uint256 public total;
//    uint256 public max;
//    uint256 public batch;
//    bool public mintable;
//    string public allocation;
//    string public uri;
//    string public uriFallback;
//    uint256 public royalty;
//    address public receiver;

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
	Max() (*big.Int, error) // Fields
	Total() (*big.Int, error)
	Batch() (*big.Int, error)
	Mintable() (bool, error)
	Allocation() (string, error)
	Uri() (string, error)
	UriFallback() (string, error)
	Royalty() (*big.Int, error)
	Receiver() (string, error)
}

var (
	_ IStorageECR721 = (*StorageECR721)(nil)
)

type StorageECR721 struct {
	session *storageerc721.BindingsSession
	ctx     context.Context
}

func (s *StorageECR721) Total() (*big.Int, error) {
	total, err := s.session.Total()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Total")
		Logger.Error(err)
		return nil, err
	}

	return total, nil
}

func (s *StorageECR721) Batch() (*big.Int, error) {
	batch, err := s.session.Batch()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Batch")
		Logger.Error(err)
		return nil, err
	}

	return batch, nil
}

func (s *StorageECR721) Mintable() (bool, error) {
	mintable, err := s.session.Mintable()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Mintable")
		Logger.Error(err)
		return false, err
	}

	return mintable, nil
}

func (s *StorageECR721) Allocation() (string, error) {
	allocation, err := s.session.Allocation()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Allocation")
		Logger.Error(err)
		return "", err
	}

	return allocation, nil
}

func (s *StorageECR721) Uri() (string, error) {
	uri, err := s.session.Uri()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "URI")
		Logger.Error(err)
		return "", err
	}

	return uri, nil
}

func (s *StorageECR721) UriFallback() (string, error) {
	uri, err := s.session.UriFallback()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "URIFallback")
		Logger.Error(err)
		return "", err
	}

	return uri, nil
}

func (s *StorageECR721) Royalty() (*big.Int, error) {
	value, err := s.session.Royalty()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Royalty")
		Logger.Error(err)
		return nil, err
	}

	return value, nil
}

func (s *StorageECR721) Receiver() (string, error) {
	value, err := s.session.Receiver()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Receiver")
		Logger.Error(err)
		return "", err
	}

	return value.String(), nil
}

func (s *StorageECR721) Max() (*big.Int, error) {
	max, err := s.session.Max()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Max")
		Logger.Error(err)
		return nil, err
	}

	return max, nil
}

func (s *StorageECR721) Mint(amount *big.Int) error {
	evmTr, err := s.session.Mint(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Mint)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", Mint, evmTr.Hash())

	return nil
}

func (s *StorageECR721) RoyaltyInfo(tokenId, salePrice *big.Int) (string, *big.Int, error) {
	address, sum, err := s.session.RoyaltyInfo(tokenId, salePrice)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", RoyaltyInfo)
		Logger.Error(err)
		return "", nil, err
	}

	return address.Hex(), sum, nil
}

func (s *StorageECR721) MintOwner(amount *big.Int) error {
	evmTr, err := s.session.MintOwner(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", MintOwner)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", MintOwner, evmTr.Hash())

	return nil
}

func (s *StorageECR721) TokenURIFallback(token *big.Int) (string, error) {
	tokenURI, err := s.session.TokenURIFallback(token)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", TokenURIFallback)
		Logger.Error(err)
		return "", err
	}

	return tokenURI, nil
}

// Price returns price
func (s *StorageECR721) Price() (int64, error) {
	price, err := s.session.Price()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", Price)
		Logger.Error(err)
		return -1, err
	}

	return price.Int64(), nil
}

// SetURI updates uri
func (s *StorageECR721) SetURI(uri string) error {
	evmTr, err := s.session.SetURI(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetURI)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetURI, evmTr.Hash())

	return nil
}

// SetAllocation updates allocation
func (s *StorageECR721) SetAllocation(allocation string) error {
	evmTr, err := s.session.SetAllocation(allocation)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetAllocation)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetAllocation, evmTr.Hash())

	return nil
}

// SetMax eth balance from token contract - setReceiver(address receiver_)
func (s *StorageECR721) SetMax(max *big.Int) error {
	evmTr, err := s.session.SetMax(max)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMax)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetMax, evmTr.Hash())

	return nil
}

// SetMintable updates mintable state
func (s *StorageECR721) SetMintable(status bool) error {
	evmTr, err := s.session.SetMintable(status)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMintable)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetMintable, evmTr.Hash())

	return nil
}

// SetRoyalty eth balance from token contract - setReceiver(address receiver_)
func (s *StorageECR721) SetRoyalty(sum *big.Int) error {
	evmTr, err := s.session.SetRoyalty(sum)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetRoyalty)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetRoyalty, evmTr.Hash())

	return nil
}

// SetReceiver eth balance from token contract - setReceiver(address receiver_)
func (s *StorageECR721) SetReceiver(receiver string) error {
	address := common.HexToAddress(receiver)

	evmTr, err := s.session.SetReceiver(address)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetReceiver)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetReceiver, evmTr.Hash())

	return nil
}

// Withdraw eth balance from token contract - withdraw()
func (s *StorageECR721) Withdraw() error {
	evmTr, err := s.session.Withdraw()
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Withdraw)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", Withdraw, evmTr.Hash())

	return nil
}
