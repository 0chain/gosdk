package znft

import (
	"context"
	"fmt"
	"os"

	"github.com/0chain/gosdk/core/logger"
	storageerc721 "github.com/0chain/gosdk/znft/contracts/dstorageerc721/binding"
	storageerc721fixed "github.com/0chain/gosdk/znft/contracts/dstorageerc721fixed/binding"
	storageerc721pack "github.com/0chain/gosdk/znft/contracts/dstorageerc721pack/binding"
	storageerc721random "github.com/0chain/gosdk/znft/contracts/dstorageerc721random/binding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

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
	Value                            int64  // Value to execute Ethereum smart contracts (default = 0)
}

func init() {
	Logger.Init(defaultLogLevel, "0chain-znft-sdk")
}

func GetConfigDir() string {
	var configDir string
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configDir = home + "/.zcn"
	return configDir
}

func (conf *Configuration) constructStorageERC721Random(address string) (*storageerc721random.Binding, *bind.TransactOpts, error) {
	storage, err := conf.createStorageERC721Random(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721RandomName)
	}

	transaction, err := conf.createTransactOpts()

	return storage, transaction, err
}

func (conf *Configuration) constructStorageERC721Pack(address string) (*storageerc721pack.Binding, *bind.TransactOpts, error) {
	storage, err := conf.createStorageERC721Pack(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721PackName)
	}

	transaction, err := conf.createTransactOpts()

	return storage, transaction, err
}

func (conf *Configuration) constructStorageERC721Fixed(address string) (*storageerc721fixed.Binding, *bind.TransactOpts, error) {
	storage, err := conf.createStorageERC721Fixed(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721FixedName)
	}

	transaction, err := conf.createTransactOpts()

	return storage, transaction, err
}

func (conf *Configuration) constructStorageERC721(address string) (*storageerc721.Binding, *bind.TransactOpts, error) {
	storage, err := conf.createStorageERC721(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721Name)
	}

	transaction, err := conf.createTransactOpts()

	return storage, transaction, err
}

func (conf *Configuration) constructWithEstimation(
	ctx context.Context,
	address string,
	method string,
	params ...interface{},
) (*storageerc721.Binding, *bind.TransactOpts, error) {
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
	abi, err := storageerc721.BindingMetaData.GetAbi()
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

func (conf *Configuration) createStorageERC721Pack(address string) (*storageerc721pack.Binding, error) {
	client, err := conf.CreateEthClient()
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721pack.NewBinding(addr, client)

	return instance, err
}

func (conf *Configuration) CreateStorageERC721PackSession(ctx context.Context, addr string) (IStorageECR721Pack, error) {
	contract, transact, err := conf.constructStorageERC721Pack(addr)
	if err != nil {
		return nil, err
	}

	session := &storageerc721pack.BindingSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &StorageECR721Pack{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}

func (conf *Configuration) CreateStorageERC721RandomSession(ctx context.Context, addr string) (IStorageECR721Random, error) {
	contract, transact, err := conf.constructStorageERC721Random(addr)
	if err != nil {
		return nil, err
	}

	session := &storageerc721random.BindingSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &StorageECR721Random{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}

func (conf *Configuration) CreateStorageERC721FixedSession(ctx context.Context, addr string) (IStorageECR721Fixed, error) {
	contract, transact, err := conf.constructStorageERC721Fixed(addr)
	if err != nil {
		return nil, err
	}

	session := &storageerc721fixed.BindingSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &StorageECR721Fixed{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}

func (conf *Configuration) CreateStorageERC721Session(ctx context.Context, addr string) (IStorageECR721, error) {
	contract, transact, err := conf.constructStorageERC721(addr)
	if err != nil {
		return nil, err
	}

	session := &storageerc721.BindingSession{
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

func (conf *Configuration) createStorageERC721(address string) (*storageerc721.Binding, error) {
	client, err := conf.CreateEthClient()
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721.NewBinding(addr, client)

	return instance, err
}

func (conf *Configuration) createStorageERC721Fixed(address string) (*storageerc721fixed.Binding, error) {
	client, err := conf.CreateEthClient()
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721fixed.NewBinding(addr, client)

	return instance, err
}

func (conf *Configuration) createStorageERC721Random(address string) (*storageerc721random.Binding, error) {
	client, err := conf.CreateEthClient()
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721random.NewBinding(addr, client)

	return instance, err
}
