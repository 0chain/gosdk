// +build js,wasm

package jsbridge

import (
	"fmt"
	"reflect"
	"syscall/js"
)

// BindFunc bind go func to js func in global
// only support
// - func(...)
// - func(...) error
// - func(...) T
// - func(...) (T,error)
func BindFunc(jsFuncName string, fn interface{}) error {

	jsFunc, err := wrappFunc(fn)
	if err != nil {
		return err
	}

	js.Global().Set(jsFuncName, jsFunc)

	return nil
}

func BindFuncs(fnList map[string]interface{}) error {

	global := js.Global()

	for jsFuncName, fn := range fnList {
		jsFunc, err := wrappFunc(fn)

		if err != nil {
			fmt.Println(err)
			return err
		}

		global.Set(jsFuncName, jsFunc)
	}

	return nil
}

func wrappFunc(fn interface{}) (js.Func, error) {
	funcType := reflect.TypeOf(fn)

	if funcType.Kind() != reflect.Func {
		return js.Func{}, ErrIsNotFunc
	}

	numOut := funcType.NumOut()

	if numOut > 2 {
		return js.Func{}, ErrFuncNotSupported
	}

	awaiter, err := Async(funcType)

	if err != nil {
		return js.Func{}, err
	}

	inputBuilder, err := NewInputBuilder(funcType).Build()

	if err != nil {
		return js.Func{}, err
	}

	invoker := reflect.ValueOf(fn)

	// jsPromise := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	resolve := args[0]
	// 	reject := args[1]

	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		in, err := inputBuilder(args)
		if err != nil {
			return js.Error{Value: js.ValueOf(err.Error())}
		}

		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			go awaiter(resolve, reject, invoker, in)

			return nil
		})

		jsFuncList = append(jsFuncList, handler)

		promise := js.Global().Get("Promise")
		return promise.New(handler)
	})

	jsFuncList = append(jsFuncList, jsFunc)

	return jsFunc, nil
}

// func InitZCNSDK(this js.Value, p []js.Value) interface{} {
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
