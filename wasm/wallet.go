package wasm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"syscall/js"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zcncore"
)

// convert JS String to []String
func strToListSring(s string) []string {
	slice := []string{}
	err := json.Unmarshal([]byte(s), &slice)

	if err != nil {
		panic(err)
	}
	return slice
}

func GetMinShardersVerify(this js.Value, p []js.Value) interface{} {
	result := zcncore.GetMinShardersVerify()
	return result
}

func GetVersion(this js.Value, p []js.Value) interface{} {
	result := zcncore.GetVersion()
	return result
}

func SetLogLevel(this js.Value, p []js.Value) interface{} {
	logLevel, _ := strconv.Atoi(p[0].String())

	zcncore.SetLogLevel(logLevel)
	return nil
}

func SetLogFile(this js.Value, p []js.Value) interface{} {
	logFile := p[0].String()
	verbose, _ := strconv.ParseBool(p[1].String())

	zcncore.SetLogFile(logFile, verbose)
	return nil
}

func CloseLog(this js.Value, p []js.Value) interface{} {
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
	sharders := strToListSring(p[1].String())
	zcncore.SetNetwork(miners, sharders)
	return nil
}

func GetNetworkJSON(this js.Value, p []js.Value) interface{} {
	result := zcncore.GetNetworkJSON()
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
	numSplits, _ := strconv.Atoi(p[1].String())
	result, err := zcncore.SplitKeys(privKey, numSplits)
	if err != nil {
		fmt.Println("error:", err)
	}
	return result
}

func strToWallet(s string) zcncrypto.Wallet {
	var w zcncrypto.Wallet
	err := json.Unmarshal([]byte(s), &w)
	if err == nil {
		fmt.Println("error:", err)
	}

	return w
}

func RegisterToMiners(this js.Value, p []js.Value) interface{} {
	s_wallet := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &WalletCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			wallet := strToWallet(s_wallet)

			err := zcncore.RegisterToMiners(&wallet, tcb)
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

func GetClientDetails(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()
	result, err := zcncore.GetClientDetails(clientID)
	if err != nil {
		fmt.Println("error:", err)
	}
	return result
}

func IsMnemonicValid(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	result := zcncore.IsMnemonicValid(mnemonic)
	return result
}

func SetWalletInfo(this js.Value, p []js.Value) interface{} {
	s_wallet := p[0].String()
	splitKeyWallet, _ := strconv.ParseBool(p[0].String())
	err := zcncore.SetWalletInfo(s_wallet, splitKeyWallet)
	if err != nil {
		fmt.Println("Cannot set wallet info")
	}
	return err
}

func SetAuthUrl(this js.Value, p []js.Value) interface{} {
	url := p[0].String()
	err := zcncore.SetAuthUrl(url)
	if err != nil {
		fmt.Println("Cannot set auth url")
	}
	return err
}

func GetBalance(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &BalanceCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetBalance(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"status": tcb.status,
				"value":  tcb.value,
				"info":   tcb.info,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetBalanceWallet(this js.Value, p []js.Value) interface{} {
	s_wallet := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &BalanceCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetBalanceWallet(s_wallet, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"status": tcb.status,
				"value":  tcb.value,
				"info":   tcb.info,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func ConvertToToken(this js.Value, p []js.Value) interface{} {
	value, _ := strconv.ParseInt(p[0].String(), 10, 64)
	result := zcncore.ConvertToToken(value)

	return result
}

func ConvertToValue(this js.Value, p []js.Value) interface{} {
	token, _ := strconv.ParseFloat(p[0].String(), 64)
	result := zcncore.ConvertToValue(token)
	return result
}

func ConvertTokenToUSD(this js.Value, p []js.Value) interface{} {
	token, _ := strconv.ParseFloat(p[0].String(), 64)
	result, _ := zcncore.ConvertTokenToUSD(token)
	return result
}

func ConvertUSDToToken(this js.Value, p []js.Value) interface{} {
	usd, _ := strconv.ParseFloat(p[0].String(), 64)
	result, _ := zcncore.ConvertUSDToToken(usd)
	return result
}

func GetLockConfig(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetLockConfig(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetLockedTokens(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetLockedTokens(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetWallet(this js.Value, p []js.Value) interface{} {
	wallet := p[0].String()
	result, err := zcncore.GetWallet(wallet)
	if err != nil {
		return err
	}
	return result
}

func GetWalletClientID(this js.Value, p []js.Value) interface{} {
	wallet := p[0].String()
	result, err := zcncore.GetWalletClientID(wallet)
	if err != nil {
		return err
	}
	return result
}

func GetZcnUSDInfo(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &USDInfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetZcnUSDInfo(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func SetupAuth(this js.Value, p []js.Value) interface{} {
	authHost := p[0].String()
	clientID := p[1].String()
	clientKey := p[2].String()
	publicKey := p[3].String()
	privateKey := p[4].String()
	localPublicKey := p[5].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &AuthCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.SetupAuth(authHost, clientID, clientKey, publicKey, privateKey, localPublicKey, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"status": tcb.status,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetIdForUrl(this js.Value, p []js.Value) interface{} {
	url := p[0].String()
	result := zcncore.GetIdForUrl(url)

	return result
}

func GetVestingPoolInfo(this js.Value, p []js.Value) interface{} {
	poolID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetVestingPoolInfo(poolID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetVestingClientList(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetVestingClientList(clientID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetVestingSCConfig(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetVestingSCConfig(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetMiners(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetMiners(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetSharders(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetSharders(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetMinerSCNodeInfo(this js.Value, p []js.Value) interface{} {
	id := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetMinerSCNodeInfo(id, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetMinerSCNodePool(this js.Value, p []js.Value) interface{} {
	id := p[0].String()
	poolID := p[1].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetMinerSCNodePool(id, poolID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetMinerSCUserInfo(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetMinerSCUserInfo(clientID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetMinerSCConfig(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetMinerSCConfig(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

// //
// // Storage SC
// //
func GetStorageSCConfig(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetStorageSCConfig(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetChallengePoolInfo(this js.Value, p []js.Value) interface{} {
	allocID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetChallengePoolInfo(allocID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetAllocation(this js.Value, p []js.Value) interface{} {
	allocID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetAllocation(allocID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetAllocations(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetAllocations(clientID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetReadPoolInfo(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetReadPoolInfo(clientID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetStakePoolInfo(this js.Value, p []js.Value) interface{} {
	blobberID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetStakePoolInfo(blobberID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetStakePoolUserInfo(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetStakePoolUserInfo(clientID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetBlobbers(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetBlobbers(tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetBlobber(this js.Value, p []js.Value) interface{} {
	blobberID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetBlobber(blobberID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetWritePoolInfo(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			tcb := &InfoCallback{}
			tcb.wg = &sync.WaitGroup{}
			tcb.wg.Add(1)

			err := zcncore.GetWritePoolInfo(clientID, tcb)
			if err != nil {
				tcb.wg.Done()
				reject.Invoke(err.Error())
			}
			tcb.wg.Wait()

			resolve.Invoke(map[string]interface{}{
				"op":     tcb.op,
				"status": tcb.status,
				"info":   tcb.info,
				"err":    tcb.err,
			})
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func Encrypt(this js.Value, p []js.Value) interface{} {
	key := p[0].String()
	text := p[0].String()

	result, err := zcncore.Encrypt(key, text)
	if err != nil {
		return err
	}
	return result
}

func Decrypt(this js.Value, p []js.Value) interface{} {
	key := p[0].String()
	text := p[0].String()

	result, err := zcncore.Decrypt(key, text)
	if err != nil {
		return err
	}
	return result
}
