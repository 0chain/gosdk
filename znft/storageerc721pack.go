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
	WithdrawToken(tokenId *big.Int) error
	TokenURI(tokenId *big.Int) (string, error)
	TokenURIFallback(tokenId *big.Int) (string, error)
	SetClosed(closed string) error
	SetOpened(opened string) error
}

var (
	_ IStorageECR721Pack = (*StorageECR721Pack)(nil)
)

type StorageECR721Pack struct {
	session *storageerc721pack.BindingsSession
	ctx     context.Context
}

func (s *StorageECR721Pack) Withdraw() error {
	evmTr, err := s.session.Withdraw0()
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Withdraw)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", Withdraw, evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) WithdrawToken(tokenId *big.Int) error {
	evmTr, err := s.session.Withdraw(tokenId)
	if err != nil {
		return err
	}

	Logger.Info("Executed %s, hash %s", "Withdraw", evmTr.Hash())

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

	Logger.Info("Executed %s, hash %s", SetReceiver, evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetRoyalty(sum *big.Int) error {
	evmTr, err := s.session.SetRoyalty(sum)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetRoyalty)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetRoyalty, evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetMintable(status bool) error {
	evmTr, err := s.session.SetMintable(status)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMintable)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetMintable, evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetMax(max *big.Int) error {
	evmTr, err := s.session.SetMax(max)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMax)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetMax, evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetAllocation(allocation string) error {
	evmTr, err := s.session.SetAllocation(allocation)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetAllocation)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetAllocation, evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetURI(uri string) error {
	evmTr, err := s.session.SetURI(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetURI)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetURI, evmTr.Hash())

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

func (s *StorageECR721Pack) SetBatch(batch *big.Int) error {
	evmTr, err := s.session.SetBatch(batch)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetBatch")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", "SetBatch", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetPrice(price *big.Int) error {
	evmTr, err := s.session.SetPrice(price)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetPrice")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", "SetPrice", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) MintOwner(amount *big.Int) error {
	evmTr, err := s.session.MintOwner(amount)
	if err != nil {
		return err
	}

	Logger.Info("Executed %s, hash %s", "MintOwner", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) Mint(amount *big.Int) error {
	evmTr, err := s.session.Mint(amount)
	if err != nil {
		return err
	}

	Logger.Info("Executed %s, hash %s", "Mint", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) Reveal(tokenId *big.Int) error {
	evmTr, err := s.session.Reveal(tokenId)
	if err != nil {
		return err
	}

	Logger.Info("Executed %s, hash %s", "Reveal", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) TokenURI(tokenId *big.Int) (string, error) {
	token, err := s.session.TokenURI(tokenId)
	if err != nil {
		return "", err
	}

	Logger.Info("Executed %s, hash %s", "TokenURI", token)

	return token, nil
}

func (s *StorageECR721Pack) TokenURIFallback(tokenId *big.Int) (string, error) {
	token, err := s.session.TokenURIFallback(tokenId)
	if err != nil {
		return "", err
	}

	Logger.Info("Executed %s, hash %s", "TokenURIFallback", token)

	return token, nil
}

func (s *StorageECR721Pack) SetClosed(closed string) error {
	evmTr, err := s.session.SetClosed(closed)
	if err != nil {
		return err
	}

	Logger.Info("Executed %s, hash %s", "SetClosed", evmTr.Hash())

	return nil
}

func (s *StorageECR721Pack) SetOpened(opened string) error {
	evmTr, err := s.session.SetOpened(opened)
	if err != nil {
		return err
	}

	Logger.Info("Executed %s, hash %s", "SetOpened", evmTr.Hash())

	return nil
}
