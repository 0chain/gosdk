//go:build js && wasm
// +build js,wasm

package indexdb

import "syscall/js"

func writeToIndexedDB(this js.Value, args []js.Value) any {
	dbName := args[0].String()
	key := args[1].String()
	value := args[2].String()

	// Open IndexedDB
	openRequest := js.Global().Get("indexedDB").Call("open", dbName)

	openRequest.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) any {
		db := openRequest.Get("result")
		tx := db.Call("transaction", "store", "readwrite")
		store := tx.Call("objectStore", "store")

		store.Call("put", value, key)
		tx.Call("oncomplete", js.FuncOf(func(this js.Value, args []js.Value) any {
			js.Global().Call("console.log", "Data written to IndexedDB successfully")
			return nil
		}))
		return nil
	}))

	openRequest.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		js.Global().Call("console.error", "Error opening IndexedDB:", openRequest.Get("error"))
		return nil
	}))

	return nil
}

func readFromIndexedDB(this js.Value, args []js.Value) any {
	dbName := args[0].String()
	key := args[1].String()

	// Open IndexedDB
	openRequest := js.Global().Get("indexedDB").Call("open", dbName)

	openRequest.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) any {
		db := openRequest.Get("result")
		tx := db.Call("transaction", "store", "readonly")
		store := tx.Call("objectStore", "store")

		getRequest := store.Call("get", key)
		getRequest.Set("onsuccess", js.FuncOf(func(this js.Value, args []js.Value) any {
			result := getRequest.Get("result")
			js.Global().Call("console.log", "Data from IndexedDB:", result)
			return nil
		}))
		getRequest.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
			js.Global().Call("console.error", "Error reading from IndexedDB:", getRequest.Get("error"))
			return nil
		}))
		return nil
	}))

	openRequest.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		js.Global().Call("console.error", "Error opening IndexedDB:", openRequest.Get("error"))
		return nil
	}))

	return nil
}
