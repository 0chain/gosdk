//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"context"

	"github.com/google/uuid"
	"github.com/hack-pad/go-webworkers/worker"
	"github.com/hack-pad/safejs"
)

type WasmWebWorker struct {
	// Name specifies an identifying name for the DedicatedWorkerGlobalScope representing the scope of the worker, which is mainly useful for debugging purposes.
	// If this is not specified, `Start` will create a UUIDv4 for it and populate back.
	Name string

	// Path is the path of the WASM to run as the Web Worker.
	// This can be a relative path on the server, or an abosolute URL.
	Path string

	// Args holds command line arguments, including the WASM as Args[0].
	// If the Args field is empty or nil, Run uses {Path}.
	Args []string

	// Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	// If Env is nil, the new Web Worker uses the current context's
	// environment.
	// If Env contains duplicate environment keys, only the last
	// value in the slice for each duplicate key is used.
	Env []string

	worker *worker.Worker
}

var (
	workers = make(map[string]*WasmWebWorker)
)

func NewWasmWebWorker(blobberID, blobberURL, clientID, publicKey, privateKey, mnemonic string) (*WasmWebWorker, error) {
	_, ok := workers[blobberID]
	if ok {
		return workers[blobberID], nil
	}

	w := &WasmWebWorker{
		Name: blobberURL,
		Env:  []string{"BLOBBER_URL=" + blobberURL, "CLIENT_ID=" + clientID, "PRIVATE_KEY=" + privateKey, "MODE=worker", "PUBLIC_KEY=" + publicKey, "MNEMONIC=" + mnemonic},
		Path: "zcn.wasm",
	}

	if err := w.Start(); err != nil {
		return nil, err
	}
	workers[blobberID] = w

	return w, nil
}

func GetWorker(blobberID string) *WasmWebWorker {
	return workers[blobberID]
}

func RemoveWorker(blobberID string) {
	worker, ok := workers[blobberID]
	if ok {
		worker.Terminate()
		delete(workers, blobberID)
	}
}

func (ww *WasmWebWorker) Start() error {
	workerJS, err := buildWorkerJS(ww.Args, ww.Env, ww.Path)
	if err != nil {
		return err
	}

	if ww.Name == "" {
		ww.Name = uuid.New().String()
	}

	wk, err := worker.NewFromScript(workerJS, worker.Options{Name: ww.Name})
	if err != nil {
		return err
	}

	ww.worker = wk

	return nil
}

// PostMessage sends data in a message to the worker, optionally transferring ownership of all items in transfers.
func (ww *WasmWebWorker) PostMessage(data safejs.Value, transfers []safejs.Value) error {
	return ww.worker.PostMessage(data, transfers)
}

// Terminate immediately terminates the Worker.
func (ww *WasmWebWorker) Terminate() {
	ww.worker.Terminate()
}

// Listen sends message events on a channel for events fired by self.postMessage() calls inside the Worker's global scope.
// Stops the listener and closes the channel when ctx is canceled.
func (ww *WasmWebWorker) Listen(ctx context.Context) (<-chan worker.MessageEvent, error) {
	return ww.worker.Listen(ctx)
}
