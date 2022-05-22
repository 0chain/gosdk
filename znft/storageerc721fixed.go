package znft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/pkg/errors"

	storageerc721fixed "github.com/0chain/gosdk/znft/contracts/dstorageerc721fixed/binding"
)

type IStorageECR721Fixed interface {
	IStorageECR721
}

type StorageECR721Fixed struct {
	session *storageerc721fixed.BindingsSession
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

func (s *StorageECR721Fixed) Withdraw() error {
	evmTr, err := s.session.Withdraw()
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Withdraw)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", Withdraw, evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) SetReceiver(receiver string) error {
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

func (s *StorageECR721Fixed) SetRoyalty(sum *big.Int) error {
	evmTr, err := s.session.SetRoyalty(sum)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetRoyalty)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetRoyalty, evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) SetMintable(status bool) error {
	evmTr, err := s.session.SetMintable(status)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMintable)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetMintable, evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) SetMax(max *big.Int) error {
	evmTr, err := s.session.SetMax(max)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetMax)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetMax, evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) SetAllocation(allocation string) error {
	evmTr, err := s.session.SetAllocation(allocation)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetAllocation)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetAllocation, evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) SetURI(uri string) error {
	evmTr, err := s.session.SetURI(uri)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", SetURI)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", SetURI, evmTr.Hash())

	return nil
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

func (s *StorageECR721Fixed) Price() (int64, error) {
	price, err := s.session.Price()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", Price)
		Logger.Error(err)
		return -1, err
	}

	return price.Int64(), nil
}

func (s *StorageECR721Fixed) Mint(amount *big.Int) error {
	evmTr, err := s.session.Mint(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", Mint)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", Mint, evmTr.Hash())

	return nil
}

func (s *StorageECR721Fixed) MintOwner(amount *big.Int) error {
	evmTr, err := s.session.MintOwner(amount)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", MintOwner)
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed %s, hash %s", MintOwner, evmTr.Hash())

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
