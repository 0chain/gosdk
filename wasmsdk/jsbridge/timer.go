//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"sync"
	"syscall/js"
	"time"
)

type Timer struct {
	sync.Mutex
	id       js.Value
	enabled  bool
	interval time.Duration
	callback func()
}

func NewTimer(interval time.Duration, callback func()) *Timer {
	return &Timer{
		interval: interval,
		callback: callback,
	}
}

func (t *Timer) Start() {
	t.Lock()
	defer t.Unlock()

	if !t.enabled {
		t.id = js.Global().Call("setInterval", js.FuncOf(t.updated), t.interval.Microseconds())
	}
}

func (t *Timer) Stop() {
	if t.enabled {
		js.Global().Call("clearInterval", t.id)
	}

	t.enabled = false
}

func (t *Timer) updated(this js.Value, args []js.Value) interface{} {
	go t.callback()
	return nil
}
