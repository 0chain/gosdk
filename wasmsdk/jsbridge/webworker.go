//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"syscall/js"

	"github.com/google/uuid"
	"github.com/hack-pad/go-webworkers/worker"
	"github.com/hack-pad/safejs"
)

const (
	MsgTypeAuth         = "auth"
	MsgTypeAuthRsp      = "auth_rsp"
	MsgTypeUpload       = "upload"
	MsgTypeUpdateWallet = "update_wallet"
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
	Env    []string
	worker *worker.Worker

	// For subscribing to events
	ctx           context.Context
	cancelContext context.CancelFunc
	subscribers   map[string]chan worker.MessageEvent
	numberOfSubs  int
	subMutex      sync.Mutex

	//isTerminated bool
	isTerminated bool
}

var (
	workers      = make(map[string]*WasmWebWorker)
	gZauthServer string
)

func NewWasmWebWorker(blobberID, blobberURL, clientID, clientKey, peerPublicKey, publicKey, privateKey, mnemonic string, isSplit bool) (*WasmWebWorker, bool, error) {
	created := false
	_, ok := workers[blobberID]
	if ok {
		return workers[blobberID], created, nil
	}

	fmt.Println("New wasm web worker, zauth server:", gZauthServer)
	w := &WasmWebWorker{
		Name: blobberURL,
		Env: []string{"BLOBBER_URL=" + blobberURL,
			"CLIENT_ID=" + clientID,
			"CLIENT_KEY=" + clientKey,
			"PEER_PUBLIC_KEY=" + peerPublicKey,
			"PRIVATE_KEY=" + privateKey,
			"MODE=worker",
			"PUBLIC_KEY=" + publicKey,
			"IS_SPLIT=" + strconv.FormatBool(isSplit),
			"MNEMONIC=" + mnemonic,
			"ZAUTH_SERVER=" + gZauthServer},
		Path:        "zcn.wasm",
		subscribers: make(map[string]chan worker.MessageEvent),
	}

	if err := w.Start(); err != nil {
		return nil, created, err
	}
	workers[blobberID] = w
	created = true
	return w, created, nil
}

func GetWorker(blobberID string) *WasmWebWorker {
	return workers[blobberID]
}

func RemoveWorker(blobberID string) {
	worker, ok := workers[blobberID]
	if ok {
		worker.subMutex.Lock()
		if worker.numberOfSubs == 0 {
			worker.Terminate()
			delete(workers, blobberID)
			worker.isTerminated = true
		}
		worker.subMutex.Unlock()
	}
}

// pass a buffered channel to subscribe to events so that the caller is not blocked
func (ww *WasmWebWorker) SubscribeToEvents(remotePath string, ch chan worker.MessageEvent) error {
	if ch == nil {
		return errors.New("channel is nil")
	}
	ww.subMutex.Lock()
	if ww.isTerminated {
		ww.subMutex.Unlock()
		return errors.New("worker is terminated")
	}
	ww.subscribers[remotePath] = ch
	ww.numberOfSubs++
	//start the worker listener if there are subscribers
	if ww.numberOfSubs == 1 {
		ctx, cancel := context.WithCancel(context.Background())
		ww.ctx = ctx
		ww.cancelContext = cancel
		eventChan, err := ww.Listen(ctx)
		if err != nil {
			ww.subMutex.Unlock()
			return err
		}
		go ww.ListenForEvents(eventChan)
	}
	ww.subMutex.Unlock()
	return nil
}

func (ww *WasmWebWorker) UnsubscribeToEvents(remotePath string) {
	ww.subMutex.Lock()
	ch, ok := ww.subscribers[remotePath]
	if ok {
		close(ch)
		delete(ww.subscribers, remotePath)
		ww.numberOfSubs--
		//stop the worker listener if there are no subscribers
		if ww.numberOfSubs == 0 {
			ww.cancelContext()
		}
	}
	ww.subMutex.Unlock()
}

func (ww *WasmWebWorker) ListenForEvents(eventChan <-chan worker.MessageEvent) {
	for {
		select {
		case <-ww.ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			//get remote path from the event
			data, err := event.Data()
			// if above throws an error, pass it to all the subscribers
			if err != nil {
				ww.removeAllSubscribers()
				return
			}
			remotePathObject, err := data.Get("remotePath")
			if err != nil {
				ww.removeAllSubscribers()
				return
			}
			remotePath, _ := remotePathObject.String()
			if remotePath == "" {
				ww.removeAllSubscribers()
				return
			}
			ww.subMutex.Lock()
			ch, ok := ww.subscribers[remotePath]
			if ok {
				ch <- event
			}
			ww.subMutex.Unlock()
		}
	}
}

func (ww *WasmWebWorker) removeAllSubscribers() {
	ww.subMutex.Lock()
	for path, ch := range ww.subscribers {
		close(ch)
		delete(ww.subscribers, path)
		ww.numberOfSubs--
	}
	ww.cancelContext()
	ww.subMutex.Unlock()
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

func SetZauthServer(zauthServer string) {
	gZauthServer = zauthServer
}

type PostWorker interface {
	PostMessage(data safejs.Value, transferables []safejs.Value) error
}

func PostMessage(w PostWorker, msgType string, data map[string]string) error {
	msgTypeUint8Array := js.Global().Get("Uint8Array").New(len(msgType))
	js.CopyBytesToJS(msgTypeUint8Array, []byte(msgType))

	obj := js.Global().Get("Object").New()
	obj.Set("msgType", msgTypeUint8Array)

	for k, v := range data {
		if k == "msgType" {
			return errors.New("msgType is key word reserved")
		}

		dataUint8Array := js.Global().Get("Uint8Array").New(len(v))
		js.CopyBytesToJS(dataUint8Array, []byte(v))
		obj.Set(k, dataUint8Array)
	}

	return w.PostMessage(safejs.Safe(obj), nil)
}

func GetMsgType(event worker.MessageEvent) (string, *safejs.Value, error) {
	data, err := event.Data()
	if err != nil {
		return "", nil, err
	}

	mt, err := data.Get("msgType")
	if err != nil {
		return "", nil, err
	}
	msgTypeLen, err := mt.Length()
	if err != nil {
		return "", nil, err
	}

	mstType := make([]byte, msgTypeLen)
	safejs.CopyBytesToGo(mstType, mt)

	return string(mstType), &data, nil
}

func SetMsgType(data *js.Value, msgType string) {
	msgTypeUint8Array := js.Global().Get("Uint8Array").New(len(msgType))
	js.CopyBytesToJS(msgTypeUint8Array, []byte(msgType))
	data.Set("msgType", msgTypeUint8Array)
}

func ParseEventDataField(data *safejs.Value, field string) (string, error) {
	fieldUint8Array, err := data.Get(field)
	if err != nil {
		return "", err
	}
	fieldLen, err := fieldUint8Array.Length()
	if err != nil {
		return "", err
	}

	fieldData := make([]byte, fieldLen)
	safejs.CopyBytesToGo(fieldData, fieldUint8Array)

	return string(fieldData), nil
}

func PostMessageToAllWorkers(msgType string, data map[string]string) error {
	for id, worker := range workers {
		fmt.Println("post message to worker", id)
		err := PostMessage(worker, msgType, data)
		if err != nil {
			return fmt.Errorf("failed to post message to worker: %s, err: %v", id, err)
		}
	}

	return nil
}
