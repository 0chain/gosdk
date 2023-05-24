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

	factoryerc721 "github.com/0chain/gosdk/znft/contracts/factorymoduleerc721/binding"
	factoryerc721fixed "github.com/0chain/gosdk/znft/contracts/factorymoduleerc721fixed/binding"
	factoryerc721pack "github.com/0chain/gosdk/znft/contracts/factorymoduleerc721pack/binding"
	factoryerc721random "github.com/0chain/gosdk/znft/contracts/factorymoduleerc721random/binding"

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
	FactoryModuleERC721PackedAddress string // FactoryModuleERC721PackedAddress address
	FactoryModuleERC721FixedAddress  string // FactoryModuleERC721FixedAddress address
	FactoryModuleERC721RandomAddress string // FactoryModuleERC721RandomAddress address
	EthereumNodeURL                  string // EthereumNodeURL URL of ethereum RPC node (infura or alchemy)
	WalletAddress                    string // WalletAddress client address
	VaultPassword                    string // VaultPassword used to sign transactions on behalf of the client
	Homedir                          string // Homedir is a client config folder
	Value                            int64  // Value to execute Ethereum smart contracts (default = 0)
}

type Znft struct {
	cfg *Configuration
}

func init() {
	Logger.Init(defaultLogLevel, "0chain-znft-sdk")
}

func NewNFTApplication(c *Configuration) *Znft {
	return &Znft{c}
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

// Functions used by session factories to create session

func (app *Znft) constructFactoryERC721(ctx context.Context, address string) (*factoryerc721.Binding, *bind.TransactOpts, error) {
	storage, err := app.createFactoryERC721(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", "FactoryERC721")
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

func (app *Znft) constructFactoryERC721Fixed(ctx context.Context, address string) (*factoryerc721fixed.Binding, *bind.TransactOpts, error) {
	storage, err := app.createFactoryERC721Fixed(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", "FactoryERC721Fixed")
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

func (app *Znft) constructFactoryERC721Pack(ctx context.Context, address string) (*factoryerc721pack.Binding, *bind.TransactOpts, error) {
	storage, err := app.createFactoryERC721Pack(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", "FactoryERC721Pack")
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

func (app *Znft) constructFactoryERC721Random(ctx context.Context, address string) (*factoryerc721random.Binding, *bind.TransactOpts, error) {
	storage, err := app.createFactoryERC721Random(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", "FactoryERC721Random")
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

func (app *Znft) constructStorageERC721Random(ctx context.Context, address string) (*storageerc721random.Binding, *bind.TransactOpts, error) {
	storage, err := app.createStorageERC721Random(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721RandomName)
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

func (app *Znft) constructStorageERC721Pack(ctx context.Context, address string) (*storageerc721pack.Binding, *bind.TransactOpts, error) {
	storage, err := app.createStorageERC721Pack(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721PackName)
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

func (app *Znft) constructStorageERC721Fixed(ctx context.Context, address string) (*storageerc721fixed.Binding, *bind.TransactOpts, error) {
	storage, err := app.createStorageERC721Fixed(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721FixedName)
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

func (app *Znft) constructStorageERC721(ctx context.Context, address string) (*storageerc721.Binding, *bind.TransactOpts, error) {
	storage, err := app.createStorageERC721(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to construct %s", ContractStorageERC721Name)
	}

	transaction, err := app.createTransactOpts(ctx)

	return storage, transaction, err
}

// Used to create StorageERC721 with preliminary gas estimation
func (app *Znft) constructWithEstimation( //nolint
	ctx context.Context,
	address string,
	method string,
	params ...interface{},
) (*storageerc721.Binding, *bind.TransactOpts, error) {
	erc721, err := app.createStorageERC721(address)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create %s in method: %s", ContractStorageERC721Name, method)
	}

	transaction, err := app.createTransactOptsWithEstimation(ctx, address, method, params)

	return erc721, transaction, err
}

// Create transaction opts with sender signature
func (app *Znft) createTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	transaction, err := app.createTransaction(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create createTransactOpts in %s", ContractStorageERC721Name)
	}

	return transaction, nil
}

func (app *Znft) createTransactOptsWithEstimation( //nolint
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

	transaction, err := app.createTransactionWithGasPrice(ctx, address, pack)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create createTransaction in %s, method: %s", ContractStorageERC721Name, method)
	}

	return transaction, nil
}

// Session factories

// Factory Sessions

func (app *Znft) CreateFactoryERC721Session(ctx context.Context, addr string) (IFactoryERC721, error) {
	contract, transact, err := app.constructFactoryERC721(ctx, addr)
	if err != nil {
		return nil, err
	}

	session := &factoryerc721.BindingSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &FactoryERC721{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}

func (app *Znft) CreateFactoryERC721PackSession(ctx context.Context, addr string) (IFactoryPack, error) {
	contract, transact, err := app.constructFactoryERC721Pack(ctx, addr)
	if err != nil {
		return nil, err
	}

	session := &factoryerc721pack.BindingSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &FactoryPack{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}

func (app *Znft) CreateFactoryERC721FixedSession(ctx context.Context, addr string) (IFactoryFixed, error) {
	contract, transact, err := app.constructFactoryERC721Fixed(ctx, addr)
	if err != nil {
		return nil, err
	}

	session := &factoryerc721fixed.BindingSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &FactoryFixed{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}

func (app *Znft) CreateFactoryERC721RandomSession(ctx context.Context, addr string) (IFactoryRandom, error) {
	contract, transact, err := app.constructFactoryERC721Random(ctx, addr)
	if err != nil {
		return nil, err
	}

	session := &factoryerc721random.BindingSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
			From:    transact.From,
			Context: ctx,
		},
		TransactOpts: *transact,
	}

	storage := &FactoryRandom{
		session: session,
		ctx:     ctx,
	}

	return storage, nil
}

// Storage Sessions

func (app *Znft) CreateStorageERC721PackSession(ctx context.Context, addr string) (IStorageECR721Pack, error) {
	contract, transact, err := app.constructStorageERC721Pack(ctx, addr)
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

func (app *Znft) CreateStorageERC721RandomSession(ctx context.Context, addr string) (IStorageECR721Random, error) {
	contract, transact, err := app.constructStorageERC721Random(ctx, addr)
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

func (app *Znft) CreateStorageERC721FixedSession(ctx context.Context, addr string) (IStorageECR721Fixed, error) {
	contract, transact, err := app.constructStorageERC721Fixed(ctx, addr)
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

func (app *Znft) CreateStorageERC721Session(ctx context.Context, addr string) (IStorageECR721, error) {
	contract, transact, err := app.constructStorageERC721(ctx, addr)
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

// Binding factories

func (app *Znft) createStorageERC721(address string) (*storageerc721.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721.NewBinding(addr, client)

	return instance, err
}

func (app *Znft) createStorageERC721Fixed(address string) (*storageerc721fixed.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721fixed.NewBinding(addr, client)

	return instance, err
}

func (app *Znft) createStorageERC721Random(address string) (*storageerc721random.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721random.NewBinding(addr, client)

	return instance, err
}

func (app *Znft) createStorageERC721Pack(address string) (*storageerc721pack.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := storageerc721pack.NewBinding(addr, client)

	return instance, err
}

func (app *Znft) createFactoryERC721(address string) (*factoryerc721.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := factoryerc721.NewBinding(addr, client)

	return instance, err
}

func (app *Znft) createFactoryERC721Pack(address string) (*factoryerc721pack.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := factoryerc721pack.NewBinding(addr, client)

	return instance, err
}

func (app *Znft) createFactoryERC721Random(address string) (*factoryerc721random.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := factoryerc721random.NewBinding(addr, client)

	return instance, err
}

func (app *Znft) createFactoryERC721Fixed(address string) (*factoryerc721fixed.Binding, error) {
	client, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(address)
	instance, err := factoryerc721fixed.NewBinding(addr, client)

	return instance, err
}
