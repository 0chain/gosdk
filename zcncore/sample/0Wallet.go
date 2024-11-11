//go:build ignore
// +build ignore

// Sample usage of the Wallet SDK - do not use.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sync"

	"github.com/0chain/gosdk/zcncore"
)

type StatusUI struct {
	i      int
	wg     sync.WaitGroup
	txnout map[string]json.RawMessage
}

// const ChainConfig = `{
// 	"miners": [
// 		"http://localhost:7071",
// 		"http://localhost:7072",
// 		"http://localhost:7073"
// 	],
// 	"sharders": [
// 		"http://localhost:7171"
// 	],
// 	"signaturescheme": "bls0chain"
// }`

const ChainConfig = `{
	"miners": [
		"http://ohio.devi.testnet-0chain.net:7071",
   		"http://ohio.devi.testnet-0chain.net:7072",
    	"http://ohio.devi.testnet-0chain.net:7073",
    	"http://ohio.devi.testnet-0chain.net:7074"
	],
	"sharders": [
		"http://ohio.devi.testnet-0chain.net:7171"
	],
	"signaturescheme": "bls0chain"
}`

var wallet = `
{"client_id":"0bc96a0980170045863d826f9eb579d8144013210602e88426408e9f83c236f6",
"client_key":"a4e58c66b072d27288b650db9a476fe66a1a4f69e0f8fb11499f9ec3a579e21e5dc0298b8c5ae5baa205730d06bc04b07a31943ab3bd620e8427c15d5c413b9e",
"keys":[
	{
		"public_key":"a4e58c66b072d27288b650db9a476fe66a1a4f69e0f8fb11499f9ec3a579e21e5dc0298b8c5ae5baa205730d06bc04b07a31943ab3bd620e8427c15d5c413b9e",
		"private_key":"c0f3a3100241888ea9c2cc5c7300e3e510a8e7190c2c20b03f80e3937a91530d"
	}],
"mnemonics":"snake mixed bird cream cotton trouble small fee finger catalog measure spoon private second canal pact unable close predict dream mask delay path inflict",
"version":"1.0",
"date_created":"2019-06-19 13:37:50.466889 -0700 PDT m=+0.023873276"}`

func (s *StatusUI) OnWalletCreateComplete(status int, w string, err string) {
	defer s.wg.Done()
	if status == zcncore.StatusError {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Wallet:", w)
	}
}

func (s *StatusUI) OnTransactionComplete(t *zcncore.Transaction, status int) {
	defer s.wg.Done()
	fmt.Println("========== TxnCompleted Status: ", status, "=======")
	fmt.Println("  Txn Hash:", t.GetTransactionHash())
}

func (s *StatusUI) OnVerifyComplete(t *zcncore.Transaction, status int) {
	defer s.wg.Done()
	fmt.Println("========== VerifyCompleted Status: ", status, "=======")
	fmt.Println(t.GetVerifyOutput())
}

func (s *StatusUI) OnBalanceAvailable(status int, value int64) {
	defer s.wg.Done()
	fmt.Println("=========== Balance Status:", status, "Value:", value,
		"Token:", zcncore.ConvertToToken(value),
		"Value:", zcncore.ConvertToValue(zcncore.ConvertToToken(value)))
}

func (zcn *StatusUI) OnAuthComplete(t *zcncore.Transaction, status int) {
	fmt.Println("Authorization complete on zauth.", status)
}

func main() {
	var cmd string
	flag.StringVar(&cmd, "cmd", "", "create|recover|validate|send|store|faucet|getbalance|verify")

	var mnemonic string
	flag.StringVar(&mnemonic, "mnemonic", "", "Mnemonic used for wallet creation.\nMandatory for -cmd recover")

	var value uint64
	flag.Uint64Var(&value, "value", 0, "Value to send")

	var txnhash string
	flag.StringVar(&txnhash, "txnhash", "", "Transaction hash to verify.\nMandatory for -cmd verify")

	var txndata string
	flag.StringVar(&txndata, "txndata", "", "Data to store.\nMandatory for -cmd store")

	var toclientID string
	flag.StringVar(&toclientID, "toclientID", "", "Receipient client ID.\nMandatory for -cmd send")

	flag.Parse()

	switch cmd {
	case "create":
	case "faucet":
	case "getbalance":
	case "recover":
		fallthrough
	case "validate":
		if mnemonic == "" {
			flag.Usage()
			return
		}
	case "send":
		if value == 0 && toclientID == "" {
			flag.Usage()
			return
		}
	case "verify":
		if txnhash == "" {
			flag.Usage()
			return
		}
	case "store":
		if txndata == "" {
			flag.Usage()
			return
		}
	default:
		fmt.Println("Unsupported command:", cmd)
		flag.Usage()
		return
	}

	err := zcncore.Init(ChainConfig)
	if err != nil {
		fmt.Println("Init failed")
		return
	}
	err = zcncore.SetWalletInfo(wallet, false)
	if err != nil {
		fmt.Println("set wallet info failed: ", err)
		return
	}

	s := &StatusUI{i: 1}
	switch cmd {
	case "create":
		s.wg.Add(1)
		err = zcncore.CreateWallet(s)
		if err != nil {
			fmt.Printf("Error create wallet: %v\n", err)
		}
		s.wg.Wait()
	case "recover":
		s.wg.Add(1)
		err = zcncore.RecoverWallet(mnemonic, s)
		if err != nil {
			fmt.Printf("Error recover wallet %v\n", err)
		}
		s.wg.Wait()
	case "validate":
		ok := zcncore.IsMnemonicValid(mnemonic)
		if ok != true {
			fmt.Println("Validate mnemonic failed")
			return
		}
		fmt.Println("**** Mnemonic is Valid ****")
	case "send":
		txn, err := zcncore.NewTransaction(s, 0, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.wg.Add(1)
		err = txn.Send(toclientID, value, "From 0Wallet sample app")
		if err != nil {
			fmt.Println("send failed: ", err)
			return
		}
		s.wg.Wait()
		s.wg.Add(1)
		txn.Verify()
		s.wg.Wait()
	case "store":
		txn, err := zcncore.NewTransaction(s, 0, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.wg.Add(1)
		err = txn.StoreData(txndata)
		if err != nil {
			fmt.Printf("store data failed: %v\n", err)
			return
		}
		s.wg.Wait()
		s.wg.Add(1)
		txn.Verify()
		s.wg.Wait()
	case "faucet":
		txn, err := zcncore.NewTransaction(s, 0, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.wg.Add(1)
		faucetAddress := "faucet smart contract address"
		methodName := "pour"
		jsonInput := "{}"
		value = 0
		err = txn.ExecuteSmartContract(faucetAddress, methodName, jsonInput, value)
		if err != nil {
			fmt.Printf("execute faucet smart contract failed: %v\n", err)
			return
		}
		s.wg.Wait()
		s.wg.Add(1)
		txn.Verify()
		s.wg.Wait()
	case "getbalance":
		s.wg.Add(1)
		err = zcncore.GetBalance(nil)
		if err != nil {
			fmt.Println("get balance failed: ", err)
			return
		}
		s.wg.Wait()
	case "getnonce":
		s.wg.Add(1)
		err = zcncore.GetNonce(nil)
		if err != nil {
			fmt.Println("get balance failed: ", err)
			return
		}
		s.wg.Wait()
	case "verify":
		txn, err := zcncore.NewTransaction(s, 0, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		txn.SetTransactionHash(txnhash)
		s.wg.Add(1)
		txn.Verify()
		s.wg.Wait()
	}
}
