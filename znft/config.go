package znft

import (
	"context"

	dstorageerc721 "github.com/0chain/gosdk/znft/contracts/dstorageerc721/binding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	ConfigFile = "config.yaml"
	WalletDir  = "wallets"
)

type Configuration struct {
	FactoryAddress                   string // FactoryAddress address
	FactoryModuleERC721Address       string // FactoryModuleERC721Address address
	FactoryModuleERC721FixedAddress  string // FactoryModuleERC721FixedAddress address
	FactoryModuleERC721RandomAddress string // FactoryModuleERC721RandomAddress address
	EthereumNodeURL                  string // EthereumNodeURL URL of ethereum RPC node (infura or alchemy)
	WalletAddress                    string // WalletAddress client address
	VaultPassword                    string // VaultPassword used to sign transactions on behalf of the client
	Homedir                          string // Homedir is a client config folder
	GasLimit                         uint64 // GasLimit limit to execute ethereum transaction
	Value                            int64  // Value to execute Ethereum smart contracts (default = 0)
}

func (conf *Configuration) constructStorageERC721(address string) (*dstorageerc721.Bindings, *bind.TransactOpts, error) {
	erc721, err := conf.createStorageERC721(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721Name)
	}

	transaction, err := conf.createTransactOpts()

	return erc721, transaction, err
}

func (conf *Configuration) constructWithEstimation(
	ctx context.Context,
	address string,
	method string,
	params ...interface{},
) (*dstorageerc721.Bindings, *bind.TransactOpts, error) {
	erc721, err := conf.createStorageERC721(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create %s in method: %s", ContractStorageERC721Name, method)
	}

	transaction, err := conf.createTransactOptsWithEstimation(ctx, address, method, params)

	return erc721, transaction, err
}

func (conf *Configuration) createTransactOpts() (*bind.TransactOpts, error) {
	transaction, err := conf.createTransaction()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create createTransactOpts in %s", ContractStorageERC721Name)
	}

	return transaction, nil
}

func (conf *Configuration) createTransactOptsWithEstimation(
	ctx context.Context,
	address, method string,
	params ...interface{},
) (*bind.TransactOpts, error) {
	// Get ABI of the contract
	abi, err := dstorageerc721.BindingsMetaData.GetAbi()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ABI in %s, method: %s", ContractStorageERC721Name, method)
	}

	// Pack the method arguments
	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pack arguments in %s, method: %s", ContractStorageERC721Name, method)
	}

	transaction, err := conf.createTransactionWithGasPrice(ctx, address, pack)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create createTransaction in %s, method: %s", ContractStorageERC721Name, method)
	}

	return transaction, nil
}

func (conf *Configuration) createStorageERC721(address string) (*dstorageerc721.Bindings, error) {
	client, err := conf.CreateEthClient()
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := dstorageerc721.NewBindings(addr, client)

	return instance, err
}

func (conf *Configuration) CreateStorageERC721Session(ctx context.Context, addr string) (IStorageECR721, error) {
	contract, transact, err := conf.constructStorageERC721(addr)
	if err != nil {
		return nil, err
	}

	session := &dstorageerc721.BindingsSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &StorageECR721{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}
