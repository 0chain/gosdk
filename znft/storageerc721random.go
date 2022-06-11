package znft

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/pkg/errors"

	storageerc721random "github.com/0chain/gosdk/znft/contracts/dstorageerc721random/binding"
)

// Solidity Functions
// function mintOwner(uint256 amount)
// function mint(uint256 amount)
// function reveal(uint256[] calldata tokens) external returns (bytes32)
// function tokenURI(uint256 tokenId) returns (string memory)
// function tokenURIFallback(uint256 tokenId) returns (string memory)
// function setHidden(string calldata hidden_)
// function setPack(address pack_)
// function setRevealable(bool status_)

type IStorageECR721Random interface {
	MintOwner(amount *big.Int) error
	Mint(amount *big.Int) error
	Reveal(tokens []*big.Int) error
	TokenURI(token *big.Int) (string, error)
	TokenURIFallback(token *big.Int) (string, error)
	SetHidden(hidden string) error
	SetPack(address common.Address) error
	SetRevealable(status bool) error
	Price() (*big.Int, error)
}

var (
	_ IStorageECR721Random = (*StorageECR721Random)(nil)
)

type StorageECR721Random struct {
	session *storageerc721random.BindingSession
	ctx     context.Context
}

func (s *StorageECR721Random) MintOwner(amount *big.Int) error {
	evmTr, err := s.session.MintOwner(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", MintOwner)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", MintOwner, " hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Random) Mint(amount *big.Int) error {
	evmTr, err := s.session.Mint(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Mint)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed ", Mint, "hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Random) Reveal(tokens []*big.Int) error {
	evmTr, err := s.session.Reveal(tokens)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "Reveal")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed Reveal, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Random) TokenURI(token *big.Int) (string, error) {
	tokenURI, err := s.session.TokenURI(token)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", TokenURI)
		Logger.Error(err)
		return "", err
	}

	return tokenURI, nil
}

func (s *StorageECR721Random) TokenURIFallback(token *big.Int) (string, error) {
	tokenURI, err := s.session.TokenURIFallback(token)
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", TokenURIFallback)
		Logger.Error(err)
		return "", err
	}

	return tokenURI, nil
}

func (s *StorageECR721Random) SetHidden(hidden string) error {
	evmTr, err := s.session.SetHidden(hidden)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetHidden")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed SetHidden, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Random) SetPack(address common.Address) error {
	evmTr, err := s.session.SetPack(address)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetPack")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed SetPack, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Random) SetRevealable(status bool) error {
	evmTr, err := s.session.SetRevealable(status)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetRevealable")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed SetRevealable, hash: ", evmTr.Hash())

	return nil
}

func (s *StorageECR721Random) Price() (*big.Int, error) {
	price, err := s.session.Price()
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "SetMax")
		Logger.Error(err)
		return price, err
	}

	return price, nil
}
