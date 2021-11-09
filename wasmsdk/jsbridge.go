// +build js,wasm

package main

import "syscall/js"

type GoFunc func(this js.Value, args []js.Value) interface{}

type JsBridge struct {
	funcs map[string]js.Func
}

// Close release resource to avoid memory leak
func (j *JsBridge) Close() {
	for _, fn := range j.funcs {
		fn.Release()
	}
}

// BindFunc bind go func to js func in global
func (j *JsBridge) BindFunc(jsFuncName string, goFunc GoFunc) *JsBridge {
	fn := js.FuncOf(goFunc)

	if j.funcs == nil {
		j.funcs = make(map[string]js.Func)
	}

	j.funcs[jsFuncName] = fn

	js.Global().Set(jsFuncName, fn)

	return j
}

// BindFuncs bind go funcs to js funcs in global
func (j *JsBridge) BindFuncs(binds map[string]GoFunc) *JsBridge {

	if j.funcs == nil {
		j.funcs = make(map[string]js.Func)
	}

	global := js.Global()

	for jsFuncName, goFunc := range binds {
		fn := js.FuncOf(goFunc)

		j.funcs[jsFuncName] = fn
		global.Set(jsFuncName, fn)
	}

	return j
}
