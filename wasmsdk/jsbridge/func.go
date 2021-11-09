// +build js,wasm

package jsbridge

type GoFunc func(b BinderBuilder) (interface{}, error)

// BindFunc bind go func to js func in global
// func BindFunc(jsFuncName string, b BinderBuilder, goFunc GoFunc) {
// 	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
// 		binder := NewBinder(args)
// 		result, err := goFunc(binder)

// 		return result

// 	})

// 	fn := js.FuncOf(goFunc)

// 	js.Global().Set(jsFuncName, fn)
// }

// blockWorker := p[0].String()
// 	signscheme := p[1].String()

// 	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
// 		resolve := args[0]
// 		reject := args[1]

// 		go func() {
// 			err := zcncore.InitZCNSDK(blockWorker, signscheme)
// 			if err != nil {
// 				reject.Invoke(err.Error())
// 			}
// 			resolve.Invoke(true)
// 		}()

// 		return nil
// 	})

// 	promiseConstructor := js.Global().Get("Promise")
// 	return promiseConstructor.New(handler)
