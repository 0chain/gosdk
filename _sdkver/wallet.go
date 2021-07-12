package main

import (
	"syscall/js"
	"encoding/json"

	"github.com/0chain/gosdk/zcncore"
)

// convert JS String to []String
func strToListSring(s string) []String {
    slice := []string{}
    err := json.Unmarshal([]byte(s), &slice)

	if err != nil {
		panic(err)
	}
	return slice
}

func GetMinShardersVerify(this js.Value) interface{} {
	result := zcncore.GetMinShardersVerify()
	return result
}

func GetVersion(this js.Value) interface{} {
	result := zcncore.GetVersion()
	return result
}

func SetLogLevel(this js.Value, p []js.Value) interface{} {
	logLevel, _  := strconv.Atoi(p[0].String())

	zcncore.SetLogFile(logLevel)
	return nil
}

func SetLogFile(this js.Value, p []js.Value) interface{} {
	logFile := p[0].String()
	verbose, _  := strconv.ParseBool(p[1].String())

	zcncore.SetLogFile(logFile, verbose)
	return nil
}

func CloseLog(this js.Value) interface{} {
	zcncore.CloseLog()
	return nil
}

func InitZCNSDK(this js.Value, p []js.Value) interface{} {
	blockWorker := p[0].String()
	signscheme := p[1].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			err := zcncore.InitZCNSDK(blockWorker, signscheme)
			if err != nil {
				reject.Invoke(err.Error())
			}
			resolve.Invoke()
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func SetNetwork(this js.Value, p []js.Value) interface{} {
	miners := strToListSring(p[0].String())
	sharders  := strToListSring(p[1].String())
	zcncore.SetNetwork(miners, sharders)
	return nil
}

func GetNetworkJSON(this js.Value, p []js.Value) interface{} {
	result = zcncore.GetNetworkJSON()
	return result
}

func CreateWallet(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &WalletCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.CreateWallet(tcb)
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

func RecoverWallet(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &WalletCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.RecoverWallet(mnemonic, tcb)
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

func SplitKeys(this js.Value, p []js.Value) interface{} {
	privKey := p[0].String()
	numSplits, _  := strconv.Atoi(p[1].String())
	result, err := zcncore.SplitKeys(privKey, numSplits)
	if err != nil {
		fmt.Println("error:", err)
	}
	return result
}