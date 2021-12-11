package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/authorizer"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
)

const (
	ConvertAmountWei = 100
)

// How should we manage nonce? - when user starts again on another server - how should we restore the value?

// Who would use this SDK

// Prerequisites:
// 1. cmd must have enough amount of Ethereum on his wallet (any Ethereum transaction will fail)
// 2. cmd must have enough WZCN tokens in Ethereum chain.

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

func main() {
	cfg := zcnbridge.ReadClientConfigFromCmd()

	if *cfg.ConfigFile == "owner" {
		runOwnerExample(cfg)
	}

	var bridge = zcnbridge.SetupBridgeClient(cfg)

	SignatureTests()

	// Testing WZCN minting side
	TraceRouteZCNToEthereumWith0ChainStab(bridge)

	// Tracing with authorizer stub executed in-proc locally. It will require 0Chain with ZCNSC SC working.
	TraceRouteZCNToEthereum(bridge)

	// Verifications of pre-performed transactions
	// Authorizers must be installed in these tests
	ConfirmEthereumTransaction()
	PrintAuthorizersList()
	PrintEthereumBurnTicketsPayloads(bridge)

	// Full test conversion
	fromERCtoZCN(bridge)
	fromZCNtoERC(bridge)
}

func runOwnerExample(cfg *zcnbridge.BridgeSDKConfig) {
	var owner = zcnbridge.SetupBridgeOwner(cfg)
	owner.AddEthereumAuthorizers(*cfg.ConfigDir)
}

// SignatureTests Create public and private keys, signs data and recovers signer public key
func SignatureTests() {
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
	data, sig := CreateSignature("message", privateKey)

	// 4. Recover public key from hashed data and signature
	pubKeyBytes := RecoverSignerPublicKey(data, sig)

	// 5. Compare
	equal := bytes.Equal(pubKeyBytes, publicKeyBytes)
	if !equal {
		fmt.Print("signatures failure")
	}
}

// CreateSignature Approach used in authorizers to sign payload to Ethereum bridge
// Returns data hash and
func CreateSignature(message string, privateKey *ecdsa.PrivateKey) (common.Hash, []byte) {
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

func RecoverSignerPublicKey(data common.Hash, signature []byte) []byte {
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
		EthereumAddress: b.GetClientEthereumWallet().Address.String(),
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
