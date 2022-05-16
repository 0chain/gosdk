package znft

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/pkg/errors"

	"github.com/0chain/gosdk/znft/contracts/dstorageerc721/binding"
	"github.com/ethereum/go-ethereum/common"
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

// SetReceiver eth balance from token contract - setReceiver(address receiver_)
func (conf *Configuration) SetReceiver(ctx context.Context, receiver string) error {
	const Method = "setReceiver"
	address := common.HexToAddress(receiver)

	erc721, transaction, err := conf.construct(ctx, Method, address)
	if err != nil {
		return errors.Wrapf(err, "failed to construct in %s", Method)
	}

	evmTr, err := erc721.SetReceiver(transaction, address)
	if err != nil {
		return errors.Wrap(err, "failed to set receiver")
	}

	evmTr.Hash()

	return nil
}

// Withdraw eth balance from token contract - withdraw()
func (conf *Configuration) Withdraw(ctx context.Context) error {
	const Method = "withdraw"

	erc721, transaction, err := conf.construct(ctx, Method)
	if err != nil {
		return errors.Wrapf(err, "failed to construct in %s", Method)
	}

	evmTr, err := erc721.Withdraw(transaction)
	if err != nil {
		return errors.Wrap(err, "failed to withdraw")
	}

	evmTr.Hash()

	return nil
}

func (conf *Configuration) construct(ctx context.Context, method string, params ...interface{}) (*binding.Bindings, *bind.TransactOpts, error) {
	erc721, err := conf.createStorageERC721()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create StorageERC721")
	}

	// Get ABI of the contract
	abi, err := binding.BindingsMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get ABI")
	}

	// Pack the method argument
	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to pack arguments")
	}

	transaction, err := conf.createTransaction(ctx, method, pack)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create createTransaction")
	}

	return erc721, transaction, nil
}

func (conf *Configuration) createStorageERC721() (*binding.Bindings, error) {
	client, err := conf.CreateEthClient()
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(conf.FactoryModuleERC721Address)

	instance, err := binding.NewBindings(addr, client)

	return instance, err
}
