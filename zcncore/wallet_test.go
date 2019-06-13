package zcncore_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/0chain/gosdk/zcncore"
)

type StatusUI struct {
	i  int
	wg sync.WaitGroup
}

type data struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
}

func (s *StatusUI) Completed(status int, w string, err string) {
	if status == zcncore.StatusError {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Wallet:", w)
	}
	s.wg.Done()
}

func TestCreateWallet(t *testing.T) {
	jsonContent := `{
		"miners": [
			"http://localhost:7071/",
			"http://localhost:7072/",
			"http://localhost:7073/"
		],
		"sharders": [
			"http://localhost:7171/"
		],
		"signaturescheme": "bls0chain"
	}`
	err := zcncore.Init(jsonContent)
	if err != nil {
		fmt.Println("Init failed")
		return
	}

	s := &StatusUI{i: 1}
	s.wg.Add(1)
	err = zcncore.CreateWallet(1, s)
	if err != nil {
		fmt.Println("Error create wallet")
	}
	s.wg.Wait()
	fmt.Println("=========Create Wallet Test Completed===========")
}
