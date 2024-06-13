//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"fmt"
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
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[recover]timer", r)
			}
		}()
		cb, _ := promise(t.updated)
		t.id = js.Global().Call("setInterval", cb, t.interval.Milliseconds())
		t.enabled = true
	}
}

func (t *Timer) Stop() {
	if t.enabled {
		js.Global().Call("clearInterval", t.id)
	}

	t.enabled = false
}

func (t *Timer) updated() {
	if t.enabled {
		t.callback()
	}
}
