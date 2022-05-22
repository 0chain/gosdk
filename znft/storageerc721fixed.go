package znft

import (
	"math/big"

	storageerc721fixed "github.com/0chain/gosdk/znft/contracts/dstorageerc721fixed/binding"
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

type IStorageECR721Fixed interface {
	IStorageECR721
}

type StorageECR721Fixed struct {
	session *storageerc721fixed.BindingsSession
}

func (s *StorageECR721Fixed) Max() (*big.Int, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Total() (*big.Int, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Batch() (*big.Int, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Mintable() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Allocation() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Uri() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) UriFallback() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Royalty() (*big.Int, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Receiver() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Withdraw() error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) SetReceiver(receiver string) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) SetRoyalty(sum *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) SetMintable(status bool) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) SetMax(max *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) SetAllocation(allocation string) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) SetURI(uri string) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) TokenURIFallback(token *big.Int) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Price() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) Mint(amount *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) MintOwner(amount *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (s StorageECR721Fixed) RoyaltyInfo(tokenId, salePrice *big.Int) (string, *big.Int, error) {
	//TODO implement me
	panic("implement me")
}
