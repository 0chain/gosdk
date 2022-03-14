package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/0chain/gosdk/zcncore"
)

//
// on JSON info available
//

type OnJSONInfoCb struct {
	value interface{}
	err   error
	wg    sync.WaitGroup
}

func (ojsonic *OnJSONInfoCb) OnInfoAvailable(op int, status int,
	info string, errMsg string) {

	defer ojsonic.wg.Done()

	if status != zcncore.StatusSuccess {
		ojsonic.err = errors.New(errMsg)
		return
	}
	if info == "" || info == "{}" {
		ojsonic.err = errors.New("empty response from sharders")
		return
	}
	var err error
	if err = json.Unmarshal([]byte(info), ojsonic.value); err != nil {
		ojsonic.err = fmt.Errorf("decoding response: %v", err)
	}
}

func (ojsonic *OnJSONInfoCb) Waiting() (err error) {
	ojsonic.wg.Wait()
	return ojsonic.err
}

func NewJSONInfoCB(val interface{}) (cb *OnJSONInfoCb) {
	cb = new(OnJSONInfoCb)
	cb.value = val
	cb.wg.Add(1)
	return
}
