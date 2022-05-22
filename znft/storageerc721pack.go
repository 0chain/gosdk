package znft

import (
	"context"

	storageerc721pack "github.com/0chain/gosdk/znft/contracts/dstorageerc721pack/binding"
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

type IStorageECR721Pack interface {
	IStorageECR721
}

type StorageECR721Pack struct {
	StorageECR721
	session *storageerc721pack.BindingsSession
	ctx     context.Context
}
