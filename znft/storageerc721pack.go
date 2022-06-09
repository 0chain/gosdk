package znft

import (
	"context"
	"math/big"

	storageerc721pack "github.com/0chain/gosdk/znft/contracts/dstorageerc721pack/binding"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// Solidity
// function mintOwner(uint256 amount)
// function mint(uint256 amount)
// function reveal(uint256 tokenId) external returns (bytes32)
// function withdraw(uint256 tokenId)
// function tokenURI(uint256 tokenId) returns (string memory)
// function tokenURIFallback(uint256 tokenId) returns (string memory)
// function setClosed(string calldata closed_)
// function setOpened(string calldata opened_)

type IStorageECR721Pack interface {
	IStorageECR721Fixed
	MintOwner(amount *big.Int) error
	Mint(amount *big.Int) error
	Reveal(tokenId *big.Int) error
	Redeem(tokenId *big.Int) error
	TokenURI(tokenId *big.Int) (string, error)
	TokenURIFallback(tokenId *big.Int) (string, error)
	SetClosed(closed string) error
	SetOpened(opened string) error
}

var (
	_ IStorageECR721Pack = (*StorageECR721Pack)(nil)
)

type StorageECR721Pack struct {
	session *storageerc721pack.BindingSession
	ctx     context.Context
}

func (s *StorageECR721Pack) SetURIFallback(uri string) error {
	evmTr, err := s.session.SetURIFallback(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetURIFallback")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed SetURIFallback, hash: ", evmTr.Hash().Hex())

	return nil
}

func (s *StorageECR721Pack) Withdraw() error {
	evmTr, err := s.session.Withdraw()
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Withdraw)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", Withdraw, " hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetReceiver(receiver string) error {
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

func (s *StorageECR721Pack) SetRoyalty(sum *big.Int) error {
	evmTr, err := s.session.SetRoyalty(sum)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetRoyalty)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetRoyalty, " hash: ", SetRoyalty, evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetMintable(status bool) error {
	evmTr, err := s.session.SetMintable(status)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMintable)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetMintable, " hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetAllocation(allocation string) error {
	evmTr, err := s.session.SetAllocation(allocation)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetAllocation)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetAllocation, " hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetURI(uri string) error {
	evmTr, err := s.session.SetURI(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetURI)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", SetURI, "hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) Price() (*big.Int, error) {
	price, err := s.session.Price()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", Price)
		Logger.Error(err)
		return nil, err
	}

	return price, nil
}

func (s *StorageECR721Pack) RoyaltyInfo(tokenId, salePrice *big.Int) (string, *big.Int, error) {
	address, sum, err := s.session.RoyaltyInfo(tokenId, salePrice)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", RoyaltyInfo)
		Logger.Error(err)
		return "", nil, err
	}

	return address.Hex(), sum, nil
}

func (s *StorageECR721Pack) Max() (*big.Int, error) {
	max, err := s.session.Max()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Max")
		Logger.Error(err)
		return nil, err
	}

	return max, nil
}

func (s *StorageECR721Pack) Total() (*big.Int, error) {
	total, err := s.session.Total()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Total")
		Logger.Error(err)
		return nil, err
	}

	return total, nil
}

func (s *StorageECR721Pack) Batch() (*big.Int, error) {
	batch, err := s.session.Batch()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Batch")
		Logger.Error(err)
		return nil, err
	}

	return batch, nil
}

func (s *StorageECR721Pack) Mintable() (bool, error) {
	mintable, err := s.session.Mintable()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Mintable")
		Logger.Error(err)
		return false, err
	}

	return mintable, nil
}

func (s *StorageECR721Pack) Allocation() (string, error) {
	allocation, err := s.session.Allocation()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Allocation")
		Logger.Error(err)
		return "", err
	}

	return allocation, nil
}

func (s *StorageECR721Pack) Uri() (string, error) {
	uri, err := s.session.Uri()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "URI")
		Logger.Error(err)
		return "", err
	}

	return uri, nil
}

func (s *StorageECR721Pack) UriFallback() (string, error) {
	uri, err := s.session.UriFallback()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "URIFallback")
		Logger.Error(err)
		return "", err
	}

	return uri, nil
}

func (s *StorageECR721Pack) Royalty() (*big.Int, error) {
	value, err := s.session.Royalty()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Royalty")
		Logger.Error(err)
		return nil, err
	}

	return value, nil
}

func (s *StorageECR721Pack) Receiver() (string, error) {
	value, err := s.session.Receiver()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Receiver")
		Logger.Error(err)
		return "", err
	}

	return value.String(), nil
}

func (s *StorageECR721Pack) MintOwner(amount *big.Int) error {
	evmTr, err := s.session.MintOwner(amount)
	if err != nil {
		return err
	}

	Logger.Info("Executed MintOwner, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) Mint(amount *big.Int) error {
	evmTr, err := s.session.Mint(amount)
	if err != nil {
		return err
	}

	Logger.Info("Executed Mint, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) Redeem(tokenId *big.Int) error {
	evmTr, err := s.session.Redeem(tokenId)
	if err != nil {
		return err
	}

	Logger.Info("Executed Reveal, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) Reveal(tokenId *big.Int) error {
	evmTr, err := s.session.Reveal(tokenId)
	if err != nil {
		return err
	}

	Logger.Info("Executed Reveal, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) TokenURI(tokenId *big.Int) (string, error) {
	token, err := s.session.TokenURI(tokenId)
	if err != nil {
		return "", err
	}

	Logger.Info("Executed TokenURI, hash: ", token)

	return token, nil
}

func (s *StorageECR721Pack) TokenURIFallback(tokenId *big.Int) (string, error) {
	token, err := s.session.TokenURIFallback(tokenId)
	if err != nil {
		return "", err
	}

	Logger.Info("Executed TokenURIFallback, hash: ", token)

	return token, nil
}

func (s *StorageECR721Pack) SetClosed(closed string) error {
	evmTr, err := s.session.SetClosed(closed)
	if err != nil {
		return err
	}

	Logger.Info("Executed SetClosed, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetOpened(opened string) error {
	evmTr, err := s.session.SetOpened(opened)
	if err != nil {
		return err
	}

	Logger.Info("Executed SetOpened, hash: ", evmTr.Hash())

	return nil
}
