package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sync"

	"0chain.net/clientsdk/zcncore"
)

type StatusUI struct {
	i  int
	wg sync.WaitGroup
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
		"http://m000.ruby.alphanet-0chain.net:7071",
		"http://m001.ruby.alphanet-0chain.net:7071",
		"http://m002.ruby.alphanet-0chain.net:7071",
		"http://m003.ruby.alphanet-0chain.net:7071",
		"http://m004.ruby.alphanet-0chain.net:7071",
		"http://m005.ruby.alphanet-0chain.net:7071",
		"http://m006.ruby.alphanet-0chain.net:7071",
		"http://m007.ruby.alphanet-0chain.net:7071",
		"http://m008.ruby.alphanet-0chain.net:7071",
		"http://m009.ruby.alphanet-0chain.net:7071"
	],
	"sharders": [
		"http://s000.ruby.alphanet-0chain.net:7171",
		"http://s001.ruby.alphanet-0chain.net:7171",
		"http://s002.ruby.alphanet-0chain.net:7171",
		"http://s003.ruby.alphanet-0chain.net:7171"
	],
	"signaturescheme": "ed25519"
}`
var wallet = `
{"client_id":"fdb9fef8c6c2e74bce76ed345ad413280ccb32f3f2a6de6894d019ac82fbfb07",
"client_key":"2221f1db1dd4c1d6f754f6681318db97cad11a15a2bc80256998926a19bcc7c5",
"keys":[
	{
		"public_key":"2221f1db1dd4c1d6f754f6681318db97cad11a15a2bc80256998926a19bcc7c5",
		"private_key":"63c1fae5d79529c2ed4e0c476bea054bc6108af47012cc8916a788810522bbbb2221f1db1dd4c1d6f754f6681318db97cad11a15a2bc80256998926a19bcc7c5"
	}],
"mnemonics":"zebra title path glory work nest admit invite tornado actress actor consider squeeze season wet elbow loop educate moment number crunch brick buzz umbrella",
"version":"1.0",
"date_created":"2019-05-13 22:56:35.317831 -0700 PDT m=+0.005299071"}`

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

func main() {
	var cmd string
	flag.StringVar(&cmd, "cmd", "", "create|recover|validate|send|store|faucet|getbalance|verify")

	var mnemonic string
	flag.StringVar(&mnemonic, "mnemonic", "", "Mnemonic used for wallet creation.\nMandatory for -cmd recover")

	var value int64
	flag.Int64Var(&value, "value", 0, "Value to send")

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
	err = zcncore.SetWalletInfo(wallet)
	if err != nil {
		fmt.Println("set wallet info failed: ", err)
		return
	}

	s := &StatusUI{i: 1}
	switch cmd {
	case "create":
		s.wg.Add(1)
		err = zcncore.CreateWallet(1, s)
		if err != nil {
			fmt.Println("Error create wallet: %v", err)
		}
		s.wg.Wait()
	case "recover":
		s.wg.Add(1)
		err = zcncore.RecoverWallet(mnemonic, 1, s)
		if err != nil {
			fmt.Println("Error recover wallet %v", err)
		}
		s.wg.Wait()
	case "validate":
		ok := zcncore.IsMnemonicValid(mnemonic)
		if ok != true {
			fmt.Println("Validate mnemonic failed")
			return
		}
		fmt.Println("\n**** Mnemonic is Valid ****\n")
	case "send":
		txn, err := zcncore.NewTransaction(s)
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
		txn, err := zcncore.NewTransaction(s)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.wg.Add(1)
		err = txn.StoreData(txndata)
		if err != nil {
			fmt.Println("store data failed: ", err)
			return
		}
		s.wg.Wait()
		s.wg.Add(1)
		txn.Verify()
		s.wg.Wait()
	case "faucet":
		txn, err := zcncore.NewTransaction(s)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.wg.Add(1)
		err = txn.ExecuteFaucetSC("pour", []byte{})
		if err != nil {
			fmt.Println("execute faucet smart contract failed: ", err)
			return
		}
		s.wg.Wait()
		s.wg.Add(1)
		txn.Verify()
		s.wg.Wait()
	case "getbalance":
		s.wg.Add(1)
		err = zcncore.GetBalance(s)
		if err != nil {
			fmt.Println("get balance failed: ", err)
			return
		}
		s.wg.Wait()
	case "verify":
		txn, err := zcncore.NewTransaction(s)
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