package zcncore

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

var (
	testKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddr    = crypto.PubkeyToAddress(testKey.PublicKey)
	testBalance = big.NewInt(2e15)
)

func TestTokensConversion(t *testing.T) {
	t.Run("Tokens to Eth", func(t *testing.T) {
		ethTokens := TokensToEth(4337488392000000000)
		require.Equal(t, 4.337488392, ethTokens)
	})

	t.Run("Eth to tokens", func(t *testing.T) {
		ethTokens := EthToTokens(4.337488392)
		require.Equal(t, int64(4337488392000000000), ethTokens)
	})

	t.Run("GTokens to Eth", func(t *testing.T) {
		ethTokens := GTokensToEth(10000000)
		require.Equal(t, 0.01, ethTokens)
	})

	t.Run("Eth to GTokens", func(t *testing.T) {
		ethTokens := GEthToTokens(4.337488392)
		require.Equal(t, int64(4337488392), ethTokens)
	})
}

func TestValidEthAddress(t *testing.T) {
	t.Run("Valid Eth wallet, but no balance", func(t *testing.T) {
		_config.chain.EthNode = "test"
		backend, _ := newTestBackend(t)
		client, _ := backend.Attach()
		defer backend.Close()
		defer client.Close()

		realClient := ethclient.NewClient(client)
		getEthClient = func() (*ethclient.Client, error) {
			return realClient, nil
		}

		res, err := IsValidEthAddress("0x531f9349ed2Fe5c526B47fa7841D35c90482e6cF")
		require.Nil(t, err, "")
		require.False(t, res, "")
	})

	t.Run("Valid Eth wallet", func(t *testing.T) {
		_config.chain.EthNode = "test"
		backend, _ := newTestBackend(t)
		client, _ := backend.Attach()
		defer backend.Close()
		defer client.Close()

		realClient := ethclient.NewClient(client)
		sendTransaction(realClient)

		getEthClient = func() (*ethclient.Client, error) {
			return realClient, nil
		}

		res, err := IsValidEthAddress(testAddr.String())
		require.Nil(t, err, "")
		require.True(t, res, "")
	})

	t.Run("Invalid Eth wallet", func(t *testing.T) {
		_config.chain.EthNode = "test"
		backend, _ := newTestBackend(t)
		client, _ := backend.Attach()
		defer backend.Close()
		defer client.Close()

		realClient := ethclient.NewClient(client)
		getEthClient = func() (*ethclient.Client, error) {
			return realClient, nil
		}

		res, err := IsValidEthAddress("testAddr.String()")
		require.NotNil(t, err, "")
		require.False(t, res, "")
	})
}

func TestGetWalletAddrFromEthMnemonic(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_config.chain.EthNode = "test"
		mnemonic := "expect domain water near beauty bag pond clap chronic chronic length leisure"
		res, err := GetWalletAddrFromEthMnemonic(mnemonic)
		require.Nil(t, err, "")
		require.Equal(t, res, "{\"ID\":\"0x531f9349ed2Fe5c526B47fa7841D35c90482e6cF\",\"PrivateKey\":\"b68bbd97a2b46b3fb2e38db771ec38935231f0b95733a4021d61601c658bc541\"}")
	})

	t.Run("Wrong", func(t *testing.T) {
		_config.chain.EthNode = "test"
		mnemonic := "this is wrong mnemonic"
		_, err := GetWalletAddrFromEthMnemonic(mnemonic)
		require.NotNil(t, err, "")
	})
}

func TestGetEthBalance(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_config.chain.EthNode = "test"
		backend, _ := newTestBackend(t)
		client, _ := backend.Attach()
		defer backend.Close()
		defer client.Close()

		realClient := ethclient.NewClient(client)
		sendTransaction(realClient)

		getEthClient = func() (*ethclient.Client, error) {
			return realClient, nil
		}

		tcb := &MockBalanceCallback{}
		tcb.wg = &sync.WaitGroup{}
		tcb.wg.Add(1)
		err := GetEthBalance(testAddr.String(), tcb)
		if err != nil {
			tcb.wg.Done()
		}
		tcb.wg.Wait()

		require.Nil(t, err, "")
		require.True(t, tcb.value > 0, "")
		require.True(t, tcb.status == StatusSuccess, "")
		require.True(t, tcb.info == "")
	})
}

func TestCheckEthHashStatus(t *testing.T) {
	t.Run("Pending transaction", func(t *testing.T) {
		_config.chain.EthNode = "test"
		backend, _ := newTestBackend(t)
		client, _ := backend.Attach()
		defer backend.Close()
		defer client.Close()

		realClient := ethclient.NewClient(client)
		sendTransaction(realClient)

		getEthClient = func() (*ethclient.Client, error) {
			return realClient, nil
		}
		result := CheckEthHashStatus("0x05aa8890d4778e292f837dd36b59a50931c175f4648c3d8157525f5454475cf7")
		require.True(t, result < 0, "")
	})
}

func TestSuggestEthGasPrice(t *testing.T) {
	t.Run("suggest gas price success", func(t *testing.T) {
		_config.chain.EthNode = "test"
		backend, _ := newTestBackend(t)
		client, _ := backend.Attach()
		defer backend.Close()
		defer client.Close()

		realClient := ethclient.NewClient(client)
		getEthClient = func() (*ethclient.Client, error) {
			return realClient, nil
		}
		gas, err := SuggestEthGasPrice()
		require.Nil(t, err)
		require.True(t, gas > 0)
	})
}

func TestTransferEthTokens(t *testing.T) {
	t.Run("success transfer", func(t *testing.T) {
		_config.chain.EthNode = "test"
		backend, _ := newTestBackend(t)
		client, _ := backend.Attach()
		defer backend.Close()
		defer client.Close()

		realClient := ethclient.NewClient(client)
		getEthClient = func() (*ethclient.Client, error) {
			return realClient, nil
		}

		hash, err := TransferEthTokens("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291", 1000000000000, 10000)
		require.Nil(t, err)
		require.EqualValues(t, hash, "0x43eba8525933e34908e766de93176d810e8582e886052708d44d9db157803aec")
	})
}

type MockBalanceCallback struct {
	wg     *sync.WaitGroup
	status int
	value  int64
	info   string
}

func (balCall *MockBalanceCallback) OnBalanceAvailable(status int, value int64, info string) {
	defer balCall.wg.Done()

	balCall.status = status
	balCall.value = value
	balCall.info = info
}

func newTestBackend(t *testing.T) (*node.Node, []*types.Block) {
	// Generate test chain.
	genesis, blocks := generateTestChain()
	// Create node
	n, err := node.New(&node.Config{})
	if err != nil {
		t.Fatalf("can't create new node: %v", err)
	}
	// Create Ethereum Service
	config := &ethconfig.Config{Genesis: genesis}
	config.Ethash.PowMode = ethash.ModeFake
	ethservice, err := eth.New(n, config)
	if err != nil {
		t.Fatalf("can't create new ethereum service: %v", err)
	}
	// Import the test chain.
	if err := n.Start(); err != nil {
		t.Fatalf("can't start test node: %v", err)
	}
	if _, err := ethservice.BlockChain().InsertChain(blocks[1:]); err != nil {
		t.Fatalf("can't import test blocks: %v", err)
	}
	return n, blocks
}

func generateTestChain() (*core.Genesis, []*types.Block) {
	db := rawdb.NewMemoryDatabase()
	config := params.AllEthashProtocolChanges
	genesis := &core.Genesis{
		Config:    config,
		Alloc:     core.GenesisAlloc{testAddr: {Balance: testBalance}},
		ExtraData: []byte("test genesis"),
		Timestamp: 9000,
	}
	// BaseFee:   big.NewInt(params.InitialBaseFee),
	generate := func(i int, g *core.BlockGen) {
		g.OffsetTime(5)
		g.SetExtra([]byte("test"))
	}
	gblock := genesis.ToBlock()
	genesis.Commit(db) //nolint: errcheck
	engine := ethash.NewFaker()
	blocks, _ := core.GenerateChain(config, gblock, engine, db, 1, generate)
	blocks = append([]*types.Block{gblock}, blocks...)
	return genesis, blocks
}

func sendTransaction(ec *ethclient.Client) error {
	// Retrieve chainID
	chainID, err := ec.ChainID(context.Background())
	if err != nil {
		return err
	}
	// Create transaction
	tx := types.NewTransaction(0, common.Address{1}, big.NewInt(1), 22000, big.NewInt(1), nil)
	signer := types.LatestSignerForChainID(chainID)
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), testKey)
	if err != nil {
		return err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return err
	}
	// Send transaction
	return ec.SendTransaction(context.Background(), signedTx)
}
