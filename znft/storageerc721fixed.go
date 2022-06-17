package znft

import (
	"context"
	"math/big"

	storageerc721fixed "github.com/0chain/gosdk/znft/contracts/dstorageerc721fixed/binding"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// Solidity Functions
// function mint(uint256 amount)
// function price() public view override returns (uint256)

type IStorageECR721Fixed interface {
	IStorageECR721
}

var (
	_ IStorageECR721Fixed = (*StorageECR721Fixed)(nil)
)

type StorageECR721Fixed struct {
	session *storageerc721fixed.BindingSession
	ctx     context.Context
}

func (s *StorageECR721Fixed) SetURIFallback(uri string) error {
	evmTr, err := s.session.SetURIFallback(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetURIFallback")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed SetURIFallback, hash: ", evmTr.Hash().Hex())

	return nil
}

func (s *StorageECR721Fixed) Total() (*big.Int, error) {
	total, err := s.session.Total()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Total")
		Logger.Error(err)
		return nil, err
	}

	return total, nil
}

func (s *StorageECR721Fixed) Batch() (*big.Int, error) {
	batch, err := s.session.Batch()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Batch")
		Logger.Error(err)
		return nil, err
	}

	return batch, nil
}

func (s *StorageECR721Fixed) Mintable() (bool, error) {
	mintable, err := s.session.Mintable()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Mintable")
		Logger.Error(err)
		return false, err
	}

	return mintable, nil
}

func (s *StorageECR721Fixed) Allocation() (string, error) {
	allocation, err := s.session.Allocation()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Allocation")
		Logger.Error(err)
		return "", err
	}

	return allocation, nil
}

func (s *StorageECR721Fixed) Uri() (string, error) {
	uri, err := s.session.Uri()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "URI")
		Logger.Error(err)
		return "", err
	}

	return uri, nil
}

func (s *StorageECR721Fixed) UriFallback() (string, error) {
	uri, err := s.session.UriFallback()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "URIFallback")
		Logger.Error(err)
		return "", err
	}

	return uri, nil
}

func (s *StorageECR721Fixed) Royalty() (*big.Int, error) {
	value, err := s.session.Royalty()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Royalty")
		Logger.Error(err)
		return nil, err
	}

	return value, nil
}

func (s *StorageECR721Fixed) Receiver() (string, error) {
	value, err := s.session.Receiver()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Receiver")
		Logger.Error(err)
		return "", err
	}

	return value.String(), nil
}

func (s *StorageECR721Fixed) Max() (*big.Int, error) {
	max, err := s.session.Max()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Max")
		Logger.Error(err)
		return nil, err
	}

	return max, nil
}

func (s *StorageECR721Fixed) Mint(amount *big.Int) error {
	evmTr, err := s.session.Mint(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Mint)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", Mint, "hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) RoyaltyInfo(tokenId, salePrice *big.Int) (string, *big.Int, error) {
	address, sum, err := s.session.RoyaltyInfo(tokenId, salePrice)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", RoyaltyInfo)
		Logger.Error(err)
		return "", nil, err
	}

	return address.Hex(), sum, nil
}

func (s *StorageECR721Fixed) MintOwner(amount *big.Int) error {
	evmTr, err := s.session.MintOwner(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", MintOwner)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", MintOwner, "hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) TokenURI(token *big.Int) (string, error) {
	tokenURI, err := s.session.TokenURI(token)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", TokenURI)
		Logger.Error(err)
		return "", err
	}

	return tokenURI, nil
}

func (s *StorageECR721Fixed) TokenURIFallback(token *big.Int) (string, error) {
	tokenURI, err := s.session.TokenURIFallback(token)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", TokenURIFallback)
		Logger.Error(err)
		return "", err
	}

	return tokenURI, nil
}

// Price returns price
func (s *StorageECR721Fixed) Price() (*big.Int, error) {
	price, err := s.session.Price()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", Price)
		Logger.Error(err)
		return nil, err
	}

	return price, nil
}

// SetURI updates uri
func (s *StorageECR721Fixed) SetURI(uri string) error {
	evmTr, err := s.session.SetURI(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetURI)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetURI, " hash: ", evmTr.Hash())

	return nil
}

// SetAllocation updates allocation
func (s *StorageECR721Fixed) SetAllocation(allocation string) error {
	evmTr, err := s.session.SetAllocation(allocation)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetAllocation)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetAllocation, "hash: ", evmTr.Hash())

	return nil
}

// SetMintable updates mintable state
func (s *StorageECR721Fixed) SetMintable(status bool) error {
	evmTr, err := s.session.SetMintable(status)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMintable)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetMintable, "hash: ", evmTr.Hash())

	return nil
}

// SetRoyalty eth balance from token contract - setReceiver(address receiver_)
func (s *StorageECR721Fixed) SetRoyalty(sum *big.Int) error {
	evmTr, err := s.session.SetRoyalty(sum)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetRoyalty)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetRoyalty, " hash: ", evmTr.Hash())

	return nil
}

// SetReceiver eth balance from token contract - setReceiver(address receiver_)
func (s *StorageECR721Fixed) SetReceiver(receiver string) error {
	address := common.HexToAddress(receiver)

	evmTr, err := s.session.SetReceiver(address)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetReceiver)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetReceiver, " hash: ", evmTr.Hash())

	return nil
}

// Withdraw eth balance from token contract - withdraw()
func (s *StorageECR721Fixed) Withdraw() error {
	evmTr, err := s.session.Withdraw()
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Withdraw)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", Withdraw, " hash: ", evmTr.Hash())

	return nil
}
