package znft

import (
	"context"
	"math/big"

	dstorageerc721 "github.com/0chain/gosdk/znft/contracts/dstorageerc721/binding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// Functions
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

const (
	Contract         = "StorageERC721"
	Withdraw         = "withdraw"
	SetReceiver      = "setReceiver"
	SetRoyalty       = "setRoyalty"
	SetMintable      = "setMintable"
	SetMax           = "setMax"
	SetAllocation    = "setAllocation"
	SetURI           = "setURI"
	TokenURIFallback = "tokenURIFallback"
	Price            = "price"
	Mint             = "mint"
	MintOwner        = "mintOwner"
	RoyaltyInfo      = "royaltyInfo"
)

// Price returns price
func (conf *Configuration) Price(ctx context.Context) error {
	session, err := conf.createStorageERC721Session(ctx, Price)
	if err != nil {
		return err
	}

	evmTr, err := session.Price()
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", Price)
	}

	evmTr.Int64()

	return nil
}

// SetURI updates uri
func (conf *Configuration) SetURI(ctx context.Context, uri string) error {
	erc721, transaction, err := conf.construct(ctx, SetURI, []byte(uri))
	if err != nil {
		return err
	}

	evmTr, err := erc721.SetURI(transaction, uri)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetURI)
	}

	evmTr.Hash()

	return nil
}

// SetAllocation updates allocation
func (conf *Configuration) SetAllocation(ctx context.Context, allocation string) error {
	erc721, transaction, err := conf.construct(ctx, SetAllocation, []byte(allocation))
	if err != nil {
		return err
	}

	evmTr, err := erc721.SetAllocation(transaction, allocation)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetAllocation)
	}

	evmTr.Hash()

	return nil
}

// SetMax eth balance from token contract - setReceiver(address receiver_)
func (conf *Configuration) SetMax(ctx context.Context, max *big.Int) error {
	erc721, transaction, err := conf.construct(ctx, SetMax, max)
	if err != nil {
		return err
	}

	evmTr, err := erc721.SetMax(transaction, max)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetMax)
	}

	evmTr.Hash()

	return nil
}

// SetMintable updates mintable state
func (conf *Configuration) SetMintable(ctx context.Context, status bool) error {
	erc721, transaction, err := conf.construct(ctx, SetMintable, status)
	if err != nil {
		return err
	}

	evmTr, err := erc721.SetMintable(transaction, status)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetMintable)
	}

	evmTr.Hash()

	return nil
}

// SetRoyalty eth balance from token contract - setReceiver(address receiver_)
func (conf *Configuration) SetRoyalty(ctx context.Context, sum *big.Int) error {
	erc721, transaction, err := conf.construct(ctx, SetRoyalty, sum)
	if err != nil {
		return err
	}

	evmTr, err := erc721.SetRoyalty(transaction, sum)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetRoyalty)
	}

	evmTr.Hash()

	return nil
}

// SetReceiver eth balance from token contract - setReceiver(address receiver_)
func (conf *Configuration) SetReceiver(ctx context.Context, receiver string) error {
	address := common.HexToAddress(receiver)

	erc721, transaction, err := conf.construct(ctx, SetReceiver, address)
	if err != nil {
		return err
	}

	evmTr, err := erc721.SetReceiver(transaction, address)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", SetReceiver)
	}

	evmTr.Hash()

	return nil
}

// Withdraw eth balance from token contract - withdraw()
func (conf *Configuration) Withdraw(ctx context.Context) error {
	erc721, transaction, err := conf.construct(ctx, Withdraw)
	if err != nil {
		return err
	}

	evmTr, err := erc721.Withdraw(transaction)
	if err != nil {
		return errors.Wrapf(err, "failed to execute %s", Withdraw)
	}

	evmTr.Hash()

	return nil
}

func (conf *Configuration) construct(ctx context.Context, method string, params ...interface{}) (*dstorageerc721.Bindings, *bind.TransactOpts, error) {
	erc721, err := conf.createStorageERC721()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create %s in method: %s", Contract, method)
	}

	transaction, err := conf.createTransactOpts(ctx, method, params)

	return erc721, transaction, err
}

func (conf *Configuration) createTransactOpts(ctx context.Context, method string, params ...interface{}) (*bind.TransactOpts, error) {
	// Get ABI of the contract
	abi, err := dstorageerc721.BindingsMetaData.GetAbi()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ABI in %s, method: %s", Contract, method)
	}

	// Pack the method arguments
	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pack arguments in %s, method: %s", Contract, method)
	}

	transaction, err := conf.createTransaction(ctx, method, pack)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create createTransaction in %s, method: %s", Contract, method)
	}

	return transaction, nil
}

func (conf *Configuration) createStorageERC721() (*dstorageerc721.Bindings, error) {
	client, err := conf.CreateEthClient()
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(conf.FactoryModuleERC721Address)
	instance, err := dstorageerc721.NewBindings(addr, client)

	return instance, err
}

func (conf *Configuration) createStorageERC721Session(ctx context.Context, method string, params ...interface{}) (*dstorageerc721.BindingsSession, error) {
	contract, transact, err := conf.construct(ctx, method, params...)
	if err != nil {
		return nil, err
	}

	session := &dstorageerc721.BindingsSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: false,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	return session, nil
}
