package znft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type IBindingsSession interface {
	Allocation() (string, error)
	BalanceOf(owner common.Address) (*big.Int, error)
	Batch() (*big.Int, error)
	Max() (*big.Int, error)
	Mintable() (bool, error)
	Owner() (common.Address, error)
	Price() (*big.Int, error)
	Receiver() (common.Address, error)
	Royalty() (*big.Int, error)
	RoyaltyInfo(tokenId *big.Int, salePrice *big.Int) (common.Address, *big.Int, error)
	Symbol() (string, error)
	TokenURI(tokenId *big.Int) (string, error)
	TokenURIFallback(tokenId *big.Int) (string, error)
	Total() (*big.Int, error)
	Uri() (string, error)
	UriFallback() (string, error)
	Mint(amount *big.Int) (*types.Transaction, error)
	MintOwner(amount *big.Int) (*types.Transaction, error)
	SetAllocation(allocation_ string) (*types.Transaction, error)
	SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error)
	SetMax(max_ *big.Int) (*types.Transaction, error)
	SetMintable(status_ bool) (*types.Transaction, error)
	SetReceiver(receiver_ common.Address) (*types.Transaction, error)
	SetRoyalty(royalty_ *big.Int) (*types.Transaction, error)
	SetURI(uri_ string) (*types.Transaction, error)
	SetURIFallback(uri_ string) (*types.Transaction, error)
	Withdraw() (*types.Transaction, error)
}
