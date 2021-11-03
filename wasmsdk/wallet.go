package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"syscall/js"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zcncore"
)

var Logger logger.Logger

func GetMinShardersVerify(this js.Value, p []js.Value) interface{} {
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			result := zcncore.GetMinShardersVerify()
			if result < 0 {
				reject.Invoke(map[string]interface{}{
					"error": "GetMinShardersVerify less than 0.",
				})
			}
			resolve.Invoke(result)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func GetVersion(this js.Value, p []js.Value) interface{} {
	result := zcncore.GetVersion()
	return result
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
			resolve.Invoke(true)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func SetWalletNetwork(this js.Value, p []js.Value) interface{} {
	var miners []string
	var sharders []string
	jsMiners := p[0]
	jsSharders := p[1]

	if got := js.Global().Get("Array").Call("isArray", jsMiners).Bool(); got {
		for i := 0; i < jsMiners.Length(); i++ {
			if got := jsMiners.Index(i).Type().String(); got == "string" {
				miners = append(miners, jsMiners.Index(i).String())
			} else {
				return map[string]interface{}{
					"error": fmt.Sprintf("SetNetwork failed. Reason: expected type \"string\". got=%#v", jsMiners.Index(i).Type().String()),
				}
			}
		}
	}

	if got := jsMiners.Type().String(); got == "string" {
		miners = append(miners, jsMiners.String())
	}

	if got := js.Global().Get("Array").Call("isArray", jsSharders).Bool(); got {
		for i := 0; i < jsSharders.Length(); i++ {
			if got := jsSharders.Index(i).Type().String(); got == "string" {
				sharders = append(sharders, jsSharders.Index(i).String())
			} else {
				return map[string]interface{}{
					"error": fmt.Sprintf("SetNetwork failed. Reason: expected type \"string\". got=%#v", jsSharders.Index(i).Type().String()),
				}
			}
		}
	}

	if got := jsSharders.Type().String(); got == "string" {
		sharders = append(sharders, jsSharders.String())
	}

	if len(miners) > 0 && len(sharders) > 0 {
		zcncore.SetNetwork(miners, sharders)
		return nil
	}

	return map[string]interface{}{
		"error": "SetNetwork failed. Reason: empty miners or sharders",
	}
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
		return map[string]interface{}{
			"error": fmt.Sprintf("SplitKeys failed. Reason: %s", err),
		}
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
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		go func() {
			result, err := zcncore.GetClientDetails(clientID)
			if err != nil {
				reject.Invoke(map[string]interface{}{
					"error": fmt.Sprintf("GetClientDetails failed. Reason: %s", err),
				})
			}

			clientDetails, err := json.Marshal(result)
			if err != nil {
				reject.Invoke(map[string]interface{}{
					"error": fmt.Sprintf("GetClientDetails to JSON Failed. Reason: %s", err),
				})
			}

			resolve.Invoke(string(clientDetails))
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

func IsMnemonicValid(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	result := zcncore.IsMnemonicValid(mnemonic)
	return result
}

func SetWalletInfo(this js.Value, p []js.Value) interface{} {
	s_wallet := p[0].String()
	splitKeyWallet := p[1].Bool()
	err := zcncore.SetWalletInfo(s_wallet, splitKeyWallet)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("SetWalletInfo failed. Reason: %s", err),
		}
	}
	return err
}

func SetAuthUrl(this js.Value, p []js.Value) interface{} {
	url := p[0].String()
	err := zcncore.SetAuthUrl(url)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("SetAuthUrl failed. Reason: %s", err),
		}
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
		return map[string]interface{}{
			"error": "Cannot get wallet",
		}
	}
	w, err := json.Marshal(result)
	if err != nil {
		return map[string]interface{}{
			"error": "Cannot Marshall wallet",
		}
	}
	return string(w)
}

func GetWalletClientID(this js.Value, p []js.Value) interface{} {
	wallet := p[0].String()
	result, err := zcncore.GetWalletClientID(wallet)
	if err != nil {
		return map[string]interface{}{
			"error": "Cannot get wallet clientId",
		}
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
	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]

		go func() {
			err := zcncore.GetIdForUrl(url)
			resolve.Invoke(err)
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
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

func GetWalletStorageSCConfig(this js.Value, p []js.Value) interface{} {
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

func GetWalletChallengePoolInfo(this js.Value, p []js.Value) interface{} {
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

func GetWalletAllocation(this js.Value, p []js.Value) interface{} {
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

func GetWalletAllocations(this js.Value, p []js.Value) interface{} {
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

func GetWalletReadPoolInfo(this js.Value, p []js.Value) interface{} {
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

func GetWalletStakePoolInfo(this js.Value, p []js.Value) interface{} {
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

func GetWalletStakePoolUserInfo(this js.Value, p []js.Value) interface{} {
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

func GetWalletBlobbers(this js.Value, p []js.Value) interface{} {
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

func GetWalletBlobber(this js.Value, p []js.Value) interface{} {
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

func GetWalletWritePoolInfo(this js.Value, p []js.Value) interface{} {
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
	text := p[1].String()

	result, err := zcncore.Encrypt(key, text)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Encrypt failed: %s", err),
		}
	}
	return result
}

func Decrypt(this js.Value, p []js.Value) interface{} {
	key := p[0].String()
	text := p[1].String()

	result, err := zcncore.Decrypt(key, text)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Decrypt failed: %s", err),
		}
	}
	return result
}
