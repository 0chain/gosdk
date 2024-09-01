//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"fmt"
	"log"
	"reflect"
	"syscall/js"
)

// BindFunc bind go func to js func in global
// only support
// - func(...)
// - func(...) error
// - func(...) T
// - func(...) (T,error)
func BindFunc(global js.Value, jsFuncName string, fn interface{}) error {

	jsFunc, err := promise(fn)
	if err != nil {
		return err
	}

	global.Set(jsFuncName, jsFunc)

	return nil
}

func BindAsyncFuncs(global js.Value, fnList map[string]interface{}) {

	for jsFuncName, fn := range fnList {
		if jsFuncName == "registerAuthorizer" || jsFuncName == "callAuth" || jsFuncName == "registerAuthCommon" {
			global.Set(jsFuncName, fn)
		} else {
			jsFunc, err := promise(fn)

			if err != nil {
				log.Println("bridge promise failed:", jsFuncName, err)
			}

			global.Set(jsFuncName, jsFunc)
		}
	}
}

func BindFuncs(global js.Value, fnList map[string]interface{}) {

	for jsFuncName, fn := range fnList {
		jsFunc, err := invoke(fn)

		if err != nil {
			log.Println("[", jsFuncName, "]", err)
			continue
		}

		global.Set(jsFuncName, jsFunc)
	}

}

func invoke(fn interface{}) (js.Func, error) {
	funcType := reflect.TypeOf(fn)

	if funcType.Kind() != reflect.Func {
		return js.Func{}, ErrIsNotFunc
	}

	numOut := funcType.NumOut()

	if numOut > 2 {
		return js.Func{}, ErrFuncNotSupported
	}

	syncInvoker, err := Sync(funcType)

	if err != nil {
		return js.Func{}, err
	}

	invoker := reflect.ValueOf(fn)

	if err != nil {
		return js.Func{}, err
	}

	inputBuilder, err := NewInputBuilder(funcType).Build()

	if err != nil {
		return js.Func{}, err
	}

	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[recover]", r)
			}
		}()

		in, err := inputBuilder(args)
		if err != nil {
			return NewJsError(err.Error())
		}

		result := syncInvoker(invoker, in)

		return result
	})

	jsFuncList = append(jsFuncList, jsFunc)

	return jsFunc, nil
}

func promise(fn interface{}) (js.Func, error) {
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

	jsFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[recover]", r)
			}
		}()

		in, err := inputBuilder(args)

		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			go awaiter(resolve, reject, invoker, in, err)

			return nil
		})

		jsFuncList = append(jsFuncList, handler)

		promise := js.Global().Get("Promise")
		return promise.New(handler)
	})

	jsFuncList = append(jsFuncList, jsFunc)

	return jsFunc, nil
}
