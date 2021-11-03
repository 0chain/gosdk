package main

import (
	"fmt"
	"strconv"
	"sync"
	"syscall/js"

	"github.com/0chain/gosdk/zcncore"
)

// JS does not have int64 so we must take a string instead of int64.
func strToInt64(s string) int64 {
	tokens, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return tokens
}

func int64ToStr(x int64) string {
	return strconv.FormatInt(x, 10)
}

func TokensToEth(this js.Value, p []js.Value) interface{} {
	tokens := strToInt64(p[0].String())
	result := zcncore.TokensToEth(tokens)
	return result
}

func EthToTokens(this js.Value, p []js.Value) interface{} {
	tokens := p[0].Float()
	result := zcncore.EthToTokens(tokens)
	return int64ToStr(result)
}

func GTokensToEth(this js.Value, p []js.Value) interface{} {
	tokens := strToInt64(p[0].String())
	result := zcncore.GTokensToEth(tokens)
	return result
}

func GEthToTokens(this js.Value, p []js.Value) interface{} {
	tokens := p[0].Float()
	result := zcncore.GEthToTokens(tokens)
	return int64ToStr(result)
}

func GetWalletAddrFromEthMnemonic(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	result, err := zcncore.GetWalletAddrFromEthMnemonic(mnemonic)
	if err != nil {
		fmt.Println("error:", err)
	}
	return result
}

func ConvertZcnTokenToETH(this js.Value, p []js.Value) interface{} {
	token := p[0].Float()
	result, err := zcncore.ConvertZcnTokenToETH(token)
	if err != nil {
		return map[string]interface{}{
			"err": err.Error(),
		}
	}
	return result
}

func SuggestEthGasPrice(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			result, err := zcncore.SuggestEthGasPrice()
			if err != nil {
				reject.Invoke(err.Error())
			}

			resolve.Invoke(int64ToStr(result))
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func TransferEthTokens(this js.Value, p []js.Value) interface{} {
	fromPrivKey := p[0].String()
	amountTokens := strToInt64(p[1].String())
	gasPrice := strToInt64(p[2].String())

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			result, err := zcncore.TransferEthTokens(fromPrivKey, amountTokens, gasPrice)
			if err != nil {
				reject.Invoke(err.Error())
			}

			resolve.Invoke(result)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

// Exports public functions in github.com/0chain/gosdk/zcncore/ethwallet.go
func IsValidEthAddress(this js.Value, p []js.Value) interface{} {
	ethAddr := p[0].String()
	success, err := zcncore.IsValidEthAddress(ethAddr)
	if err != nil {
		return map[string]interface{}{
			"err": err.Error(),
		}
	}
	return success
}

func GetEthBalance(this js.Value, p []js.Value) interface{} {
	ethAddress := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &BalanceCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetEthBalance(ethAddress, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"info":   tcb.info,
				"status": tcb.status,
				"value":  tcb.value,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func CreateWalletFromEthMnemonic(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	password := p[1].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &WalletCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.CreateWalletFromEthMnemonic(mnemonic, password, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"status": tcb.status,
				"err":    tcb.err,
				"wallet": tcb.wallet,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

type WalletCallback struct {
	wg     *sync.WaitGroup
	status int
	wallet string
	err    string
}

type BalanceCallback struct {
	wg     *sync.WaitGroup
	status int
	value  int64
	info   string
}

type InfoCallback struct {
	wg     *sync.WaitGroup
	op     int
	status int
	info   string
	err    string
}

type USDInfoCallback struct {
	wg     *sync.WaitGroup
	status int
	info   string
	err    string
}

type AuthCallback struct {
	wg     *sync.WaitGroup
	status int
	err    string
}

func (balCall *BalanceCallback) OnBalanceAvailable(status int, value int64, info string) {
	defer balCall.wg.Done()

	balCall.status = status
	balCall.value = value
	balCall.info = info
}

func (wallCall *WalletCallback) OnWalletCreateComplete(status int, w string, err string) {
	defer wallCall.wg.Done()
	fmt.Println("callback [status, w, err]:", status, w, err)

	wallCall.status = status
	wallCall.wallet = w
	wallCall.err = err
}

func (infoCall *InfoCallback) OnInfoAvailable(op int, status int, info string, err string) {
	defer infoCall.wg.Done()
	fmt.Println("callback [op, status, info, err]:", op, status, info, err)

	infoCall.op = op
	infoCall.status = status
	infoCall.info = info
	infoCall.err = err
}

func (usdCall *USDInfoCallback) OnUSDInfoAvailable(status int, info string, err string) {
	defer usdCall.wg.Done()
	fmt.Println("callback [status, info, err]:", status, info, err)

	usdCall.status = status
	usdCall.info = info
	usdCall.err = err
}

func (authCall *AuthCallback) OnSetupComplete(status int, err string) {
	defer authCall.wg.Done()
	fmt.Println("callback [status, err]:", status, err)

	authCall.status = status
	authCall.err = err
}
