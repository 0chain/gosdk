package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"path"
	"time"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/authorizer"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	hdw "github.com/miguelmota/go-ethereum-hdwallet"
	"go.uber.org/zap"
)

const (
	ConvertAmountWei = 100
)

// ? How should we manage nonce ?
// When user starts again on another server - how should we restore the nonce value?

// Prerequisites:
// 1. cmd must have enough amount of Ethereum on his wallet (any Ethereum transaction will fail)
// 2. cmd must have enough WZCN tokens in Ethereum chain.
// 3. Address of the client should be initializer in storage using ImportAccount(mnemonic, password)

// main:
// `--config_file bridge` runs bridge client
// `--config_file owner`  runs owner client
func main() {
	zcnbridge.CreateInitialClientConfig(
		"bridge.yaml",
		"0x860FA46F170a87dF44D7bB867AA4a5D2813127c1",
		"0xF26B52df8c6D9b9C20bfD7819Bed75a75258c7dB",
		"0x930E1BE76461587969Cb7eB9BFe61166b1E70244",
		"https://ropsten.infura.io/v3/22cb2849f5f74b8599f3dc2a23085bd4",
		"password",
		300000,
		0,
		75.0,
	)

	zcnbridge.CreateInitialOwnerConfig(
		"owner.yaml",
		"0x860FA46F170a87dF44D7bB867AA4a5D2813127c1",
		"0xF26B52df8c6D9b9C20bfD7819Bed75a75258c7dB",
		"0x930E1BE76461587969Cb7eB9BFe61166b1E70244",
		"0xFE20Ce9fBe514397427d20C91CB657a4478A0FFa",
		"https://ropsten.infura.io/v3/22cb2849f5f74b8599f3dc2a23085bd4",
		"password",
		300000,
		0,
	)

	// First is read config from command line
	cfg := zcnbridge.ReadClientConfigFromCmd()

	// Checking if an account exists in key storage
	if zcnbridge.AccountExists("0x860FA46F170a87dF44D7bB867AA4a5D2813127c1") {
		fmt.Println("Account exists")
	}

	// List all accounts initialized in storage
	zcnbridge.ListStorageAccounts()

	// Next step is register your account in the key storage if it doesn't exist (mandatory)
	// This should be done in zwallet cli
	registerAccountInKeyStorage(
		"tag volcano eight thank tide danger coast health above argue embrace heavy",
		"password",
	)

	// How to manage key storage example sets
	keyStorageExample()
	// How to sign using legacy and dynamic transactions
	signingExamples()

	// Owner examples: adding new authorizer
	if *cfg.ConfigFile == "owner" {
		runBridgeOwnerExample(cfg)
		return
	}

	// Bridge client examples
	runBridgeClientExample(cfg)
}

func registerAccountInKeyStorage(mnemonic, password string) {
	err := zcnbridge.ImportAccount(mnemonic, password)
	if err != nil {
		fmt.Println(err)
	}
}

// keyStorageExample Shows how new/existing user will work with key storage
// 1. If user is new, user creates new storage with key and/or mnemonic
// 2. if user exists, user can sign transactions using public key and password to unlock the key storage
func keyStorageExample() {
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	password := "password"

	createKeyStorage(password, true)
	importFromMnemonicToStorage(mnemonic, password, false)
	signWithKeyStore("0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947", password)
}

func signingExamples() {
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	signLegacyTransactionExample(mnemonic)
	signDynamicTransactorExample(mnemonic)
}

// createKeyStorage create new key storage and a new account
func createKeyStorage(password string, delete bool) {
	keyDir := path.Join(zcnbridge.GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(account.Address.Hex())

	// 5. Delete key store

	if delete {
		err = ks.Delete(account, password)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// signWithKeyStore signs the transaction using public key and key storage
func signWithKeyStore(address, password string) {
	// 1. Create storage and account if it doesn't exist and add account to it

	keyDir := path.Join(zcnbridge.GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// Create account definitions
	fromAccDef := accounts.Account{
		Address: common.HexToAddress(address),
	}

	// Find the signing account
	signAcc, err := ks.Find(fromAccDef)
	if err != nil {
		fmt.Printf("account keystore find error %v", err)
		return
	}

	// Unlock the signing account
	errUnlock := ks.Unlock(signAcc, password)
	if errUnlock != nil {
		fmt.Println("account unlock error:")
		return
	}
	fmt.Printf("account unlocked: signAcc.addr=%s; signAcc.url=%s\n", signAcc.Address.String(), signAcc.URL)

	nonce := uint64(0)
	chainID := big.NewInt(3)
	value := big.NewInt(1000000000000000000)
	toAddress := common.HexToAddress("0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	var data []byte

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	signedTx, err := ks.SignTx(signAcc, tx, chainID)
	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(signedTx)
}

// importFromMnemonicToStorage Importing wallet to key storage from mnemonic
func importFromMnemonicToStorage(mnemonic, password string, delete bool) {
	// 1. Create storage and account if it doesn't exist and add account to it

	keyDir := path.Join(zcnbridge.GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// 2. Init wallet

	wallet, err := hdw.NewFromMnemonic(mnemonic)
	if err != nil {
		fmt.Println(err)
		return
	}

	pathD := hdw.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(pathD, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	key, err := wallet.PrivateKey(account)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 3. Find key

	acc, err := ks.Find(account)
	if err == nil {
		fmt.Printf("Account found: %s\n", acc.Address.Hex())
		fmt.Println(acc.URL.Path)
		return
	}

	// 4. Import the key if it doesn't exist

	acc, err = ks.ImportECDSA(key, password)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(acc.URL.Path)
	fmt.Println(acc.Address.Hex())
	fmt.Println(account.URL.Path)
	fmt.Println(account.Address.Hex())

	// 5. Delete key store

	if delete {
		err = ks.Delete(account, password)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// signDynamicTransactorExample - London hard fork, is set when gasPrice is set nil
func signDynamicTransactorExample(mnemonic string) {
	ctx := context.Background()
	client, err := ethclient.Dial("https://ropsten.infura.io/v3/22cb2849f5f74b8599f3dc2a23085bd4")
	chainID, _ := client.ChainID(ctx)

	wallet, err := hdw.NewFromMnemonic(mnemonic)
	if err != nil {
		fmt.Println(err)
		return
	}

	pathD := hdw.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(pathD, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	key, err := wallet.PrivateKey(account)

	signer := types.NewLondonSigner(chainID)

	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	head, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if head.BaseFee == nil {
		return
	}

	gasFeeCap := new(big.Int).Add(
		gasTipCap,
		new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
	)

	nonce, err := client.PendingNonceAt(ctx, account.Address)
	if err != nil {
		fmt.Println(err)
		return
	}

	value := big.NewInt(1000000000000000000)
	toAddress := common.HexToAddress("0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947")
	gasLimit := uint64(21000)
	var data []byte

	baseTx := &types.DynamicFeeTx{
		To:        &toAddress,
		Nonce:     nonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       gasLimit,
		Value:     value,
		Data:      data,
	}

	tx := types.NewTx(baseTx)

	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
	if err != nil {
		zcnbridge.ExitWithError(err)
	}

	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(signedTx)

	// Sending raw transaction
	//_ := client.SendTransaction(ctx, signedTx)
}

// signLegacyTransactionExample is called when gas price is set
func signLegacyTransactionExample(mnemonic string) {
	wallet, err := hdw.NewFromMnemonic(mnemonic)
	if err != nil {
		fmt.Println(err)
		return
	}

	//wallet.SetFixIssue172(true)

	pathD := hdw.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(pathD, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(account.Address.Hex())

	url, err := wallet.Path(account)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(url)

	lena := wallet.Accounts()
	fmt.Println(lena)

	// Signing

	nonce := uint64(0)
	value := big.NewInt(1000000000000000000)
	toAddress := common.HexToAddress("0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	var data []byte

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	spew.Dump(signedTx)

	// Sending raw transaction
	//_ := client.SendTransaction(ctx, transaction)
}

func runBridgeClientExample(cfg *zcnbridge.BridgeSDKConfig) {
	var bridge = zcnbridge.SetupBridgeClientSDK(cfg)

	balance, err := bridge.GetBalance()
	if err == nil {
		fmt.Println(balance)
	}

	signatureTests()

	// Full test conversion
	fromERCtoZCN(bridge)
	fromZCNtoERC(bridge)

	// Testing WZCN minting side
	TraceRouteZCNToEthereumWith0ChainStab(bridge)

	// Tracing with authorizer stub executed in-proc locally. It will require 0Chain with ZCNSC SC working.
	TraceRouteZCNToEthereum(bridge)

	// Verifications of pre-performed transactions
	// Authorizers must be installed in these tests
	ConfirmEthereumTransaction()
	PrintAuthorizersList()
	PrintEthereumBurnTicketsPayloads(bridge)
}

func runBridgeOwnerExample(cfg *zcnbridge.BridgeSDKConfig) {
	var owner = zcnbridge.SetupBridgeOwnerSDK(cfg)
	owner.AddEthereumAuthorizers(*cfg.ConfigDir)
}

// signatureTests Create public and private keys, signs data and recovers signer public key
func signatureTests() {
	// 1. Private Key
	privateKeyHex := "fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19"
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		fmt.Print(err)
	}

	// 2. Public Key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Print(err)
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	// 3. Create data and signature
	data, sig := createSignature("message", privateKey)

	// 4. Recover public key from hashed data and signature
	pubKeyBytes := recoverSignerPublicKey(data, sig)

	// 5. Compare
	equal := bytes.Equal(pubKeyBytes, publicKeyBytes)
	if !equal {
		fmt.Print("signatures failure")
	}
}

// createSignature Approach used in authorizers to sign payload to Ethereum bridge
// Returns data hash and
func createSignature(message string, privateKey *ecdsa.PrivateKey) (common.Hash, []byte) {
	// 1. Hash data
	data := []byte(message)
	hash := crypto.Keccak256Hash(data)
	fmt.Println(hash.Hex())

	// 2. Signing the data
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		fmt.Print(err)
	}

	// 3. Signature to hex string
	fmt.Println("Signature length: ", len(signature))
	sig := hexutil.Encode(signature)
	fmt.Println(sig)

	return hash, signature
}

func recoverSignerPublicKey(data common.Hash, signature []byte) []byte {
	sigPublicKey, err := crypto.Ecrecover(data.Bytes(), signature)
	if err != nil {
		fmt.Println(err)
	}

	return sigPublicKey
}

// TraceRouteZCNToEthereumWith0ChainStab Implements to WZCN Ethereum minting
// It will use ZCNSC SC Burn stab and won't require 0Chain working.
// It's possible to test WZCN burning in Ethereum side without 0Chain working
func TraceRouteZCNToEthereumWith0ChainStab(b *zcnbridge.BridgeClient) {
	output := GenerateBurnTransactionOutput(b)
	TraceEthereumMint(b, string(output))
}

// TraceRouteZCNToEthereum Traces the route from ZCN burn to Ethereum mint, bypassing authorizer part which
// was duplicated here
func TraceRouteZCNToEthereum(b *zcnbridge.BridgeClient) {
	// --------------------- This part is executed in client in GOSDK part -------------------------------
	// Sends {nonce,ethereum_address} payload to burn function
	tx, err := b.BurnZCN(context.TODO(), ConvertAmountWei)
	if err != nil {
		fmt.Print(err)
		return
	}

	tx, err = b.VerifyZCNTransaction(context.TODO(), tx.Hash)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("Burn transaction hash: %s\n. Confirmed", tx.Hash)
	fmt.Printf("Burn transaction output: %s\n", tx.TransactionOutput)
	// ---------------------- End SDK -------------------------------------------------------------------

	output := tx.TransactionOutput

	// ---------------------  This part is executed in authorizers in /burnticket handler ---------------
	TraceEthereumMint(b, output)
}

var nonce int64

// GenerateBurnTransactionOutput stub for burn transaction in ZCN Chain
func GenerateBurnTransactionOutput(b *zcnbridge.BridgeClient) []byte {
	// Type of input of burn transaction
	type BurnPayload struct {
		Nonce           int64  `json:"nonce"`
		EthereumAddress string `json:"ethereum_address"`
	}

	// Type of response of burn transaction
	type BurnPayloadResponse struct {
		TxnID           string `json:"0chain_txn_id"`
		Nonce           int64  `json:"nonce"`
		Amount          int64  `json:"amount"`
		EthereumAddress string `json:"ethereum_address"`
	}

	var scAddress = "6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712e0"
	nonce++

	// Executed at burn function in smartcontract
	payload := &BurnPayload{
		Nonce:           nonce,
		EthereumAddress: b.EthereumAddress,
	}

	// generating transaction hash
	transactionData, _ := json.Marshal(payload)
	hashData := fmt.Sprintf("%v:%v:%v:%v:%v", time.Now(), b.ID(), scAddress, 0, encryption.Hash(transactionData))
	hash := encryption.Hash(hashData)

	output := &BurnPayloadResponse{
		TxnID:           hash,
		Nonce:           payload.Nonce,
		Amount:          ConvertAmountWei,
		EthereumAddress: payload.EthereumAddress,
	}

	buffer, _ := json.Marshal(output)

	return buffer
}

func TraceEthereumMint(b *zcnbridge.BridgeClient, output string) {
	// --------------------- This part is executed in authorizers in /burnticket handler ---------------
	// Sends hash to authorizer
	pb := &authorizer.ProofOfBurn{}
	err := pb.Decode([]byte(output)) // TODO: Ensure that transaction output contains proofOfBurn
	if err != nil {
		fmt.Print(err)
	}
	err = pb.Verify()
	if err != nil {
		fmt.Print(err)
		return
	}

	err = pb.Sign(b)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("Generated signature of length: %d\n", len(pb.Signature))

	buf := bytes.NewBuffer(nil)
	_ = json.NewEncoder(buf).Encode(pb)
	bridgeAnswer := buf.Bytes()

	// --------------- END authorizer --------------------------------

	// --------------- GOSDK part executed in Client -----------------
	burnTicket := &zcnbridge.ProofZCNBurn{}
	err = json.Unmarshal(bridgeAnswer, burnTicket)
	if err != nil {
		fmt.Print(err)
		return
	}

	var sigs []*ethereum.AuthorizerSignature

	sig := &ethereum.AuthorizerSignature{
		ID:        burnTicket.GetAuthorizerID(),
		Signature: burnTicket.Signature,
	}

	sigs = append(sigs, sig)

	payload := &ethereum.MintPayload{
		ZCNTxnID:   burnTicket.TxnID,
		Amount:     burnTicket.Amount,
		Nonce:      burnTicket.Nonce,
		Signatures: sigs,
	}
	// --------------- END GOSDK part executed in Client -----------------------------

	// --------------- GOSDK part starts on the client and executed in Ethereum ------
	ethTrx, err := b.MintWZCN(context.TODO(), payload)
	if err != nil {
		fmt.Print(err)
		return
	}

	status, err := zcnbridge.ConfirmEthereumTransaction(ethTrx.Hash().String(), 5, time.Second)
	if err != nil {
		fmt.Print(err)
		return
	}

	if status == 1 {
		fmt.Println("Transaction is successful")
	} else {
		fmt.Println("Transaction failed")
	}

	// ---------------- Completed ZCN -> WZCN transaction ------------------------------
}

// Ropsten burn successful transactions for which we may receive burn tickets and mint payloads
// to mint ZCN tokens
var tranHashes = []string{
	"0xa5049192c3622534e6195fbadcf21c9eb928ca3e5e8c7056f500f78f31c1c1aa",
	"0xd3583513ea4f76f25000e704c8fc12c5b7b71a1574138d4df20d948255bd7f9c",
	"0x468805e8bb268d584659ccd104e36bd5e552feec440d1a761aa8f9034a92b2fd",
	"0x39ba7befd88a6dc6abec1bd503a6c2ced9472b8643704e4048d673728fb373b5",
	"0x31925839586949a96e72cacf25fed7f47de5faff78adc20946183daf3c4cf230",
	"0xef7494153ca9ddb871f4ca385ebaf47c572fbe14c39f98b5decc6d91b9230dd3",
	"0x943f86ca64a87adc346bc46a6732ea4a4c0eb7dee1453b1c37fb86f144f88658",
	"0x29ce974e8a44e6628af4749d50df04b6555bd3b932f080b0447bbe4d61f09a90",
	"0xe0c3941fc74ea7e17a80750e5923e2fca8e7db3dcf9b67d2ab4e1528524fe808",
	"0x5f8efdce13d0235c273b3714bcad8817cacb6d60867b156032f3e52cd6f32ebe",
}

func ConfirmEthereumTransaction() {
	for _, hash := range tranHashes {
		status, err := zcnbridge.ConfirmEthereumTransaction(hash, 10, time.Second)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Ttansaction %s status: %d\n", hash, status)
	}
}

func PrintEthereumBurnTicketsPayloads(b *zcnbridge.BridgeClient) {
	for _, hash := range tranHashes {
		payload, err := b.QueryZChainMintPayload(hash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(payload)
	}
}

func PrintAuthorizersList() {
	authorizers, err := zcnbridge.GetAuthorizers()
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(authorizers)
}

func fromZCNtoERC(b *zcnbridge.BridgeClient) {
	burnTrx, err := b.BurnZCN(context.TODO(), ConvertAmountWei)
	burnTrxHash := burnTrx.Hash
	if err != nil {
		log.Logger.Fatal("failed to burn in ZCN", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	burnTrx, err = b.VerifyZCNTransaction(context.TODO(), burnTrxHash)
	if err != nil {
		return
	}

	// ASK authorizers for burn tickets to mint in Ethereum
	mintPayload, err := b.QueryEthereumMintPayload(burnTrxHash)
	if err != nil {
		log.Logger.Fatal("failed to verify burn transactions in ZCN in QueryEthereumMintPayload", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	mintTrx, err := b.MintWZCN(context.Background(), mintPayload)
	tranHash := mintTrx.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute MintWZCN", zap.Error(err), zap.String("hash", tranHash))
	}

	// ASK for minting events from bridge contract but this is not necessary as we're going to check it by hash

	res, err := zcnbridge.ConfirmEthereumTransaction(tranHash, 60, time.Second)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransaction",
			zap.String("hash", tranHash),
			zap.Error(err),
		)
	}

	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction ConfirmEthereumTransaction", zap.String("hash", tranHash))
	}
}

func fromERCtoZCN(b *zcnbridge.BridgeClient) {
	// Example: https://ropsten.etherscan.io/tx/0xa28266fb44cfc2aa27b26bd94e268e40d065a05b1a8e6339865f826557ff9f0e
	transaction, err := b.IncreaseBurnerAllowance(context.Background(), ConvertAmountWei)
	if err != nil {
		log.Logger.Fatal("failed to execute IncreaseBurnerAllowance", zap.Error(err))
	}

	hash := transaction.Hash().Hex()
	res, err := zcnbridge.ConfirmEthereumTransaction(hash, 60, time.Second)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransaction",
			zap.String("hash", hash),
			zap.Error(err),
		)
	}
	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction", zap.String("hash", transaction.Hash().Hex()))
	}

	burnTrx, err := b.BurnWZCN(context.Background(), ConvertAmountWei)
	burnTrxHash := burnTrx.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute BurnWZCN in wrapped chain", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	res, err = zcnbridge.ConfirmEthereumTransaction(burnTrxHash, 60, time.Second)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransaction",
			zap.String("hash", burnTrxHash),
			zap.Error(err),
		)
	}
	if res == 0 {
		log.Logger.Fatal("failed to confirm burn transaction in ZCN in ConfirmEthereumTransaction", zap.String("hash", burnTrxHash))
	}

	// ASK authorizers for burn tickets to mint in WZCN
	mintPayload, err := b.QueryZChainMintPayload(burnTrxHash)
	if err != nil {
		log.Logger.Fatal("failed to QueryZChainMintPayload", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	mintTrx, err := b.MintZCN(context.TODO(), mintPayload)
	if err != nil {
		log.Logger.Fatal("failed to MintZCN", zap.Error(err), zap.String("hash", mintTrx.Hash))
	}
}
