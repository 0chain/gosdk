package zcncore

import (
	"encoding/json"
	stderrors "errors"

	thrown "github.com/0chain/errors"
)

// GetUserLockedTotal get total token user locked
// # Inputs
//   - clientID wallet id
func GetUserLockedTotal(clientID string) (int64, error) {

	err := checkSdkInit()
	if err != nil {
		return 0, err
	}

	var url = withParams(STORAGESC_GET_USER_LOCKED_TOTAL, Params{
		"client_id": clientID,
	})
	cb := createGetInfoCallback()
	go GetInfoFromSharders(url, OpStorageSCGetStakePoolInfo, cb)
	info, err := cb.Wait()
	if err != nil {
		return 0, err
	}

	result := make(map[string]int64)

	err = json.Unmarshal([]byte(info), &result)
	if err != nil {
		return 0, thrown.Throw(err, "invalid json format")
	}

	total, ok := result["total"]
	if ok {
		return total, nil
	}

	return 0, stderrors.New("invalid result")

}

func createGetInfoCallback() *getInfoCallback {
	return &getInfoCallback{
		callback: make(chan bool),
	}
}

type getInfoCallback struct {
	callback chan bool
	status   int
	info     string
	err      string
}

func (cb *getInfoCallback) OnInfoAvailable(op int, status int, info string, err string) {

	// if status == StatusSuccess then info is valid
	// is status != StatusSuccess then err will give the reason

	cb.status = status
	if status == StatusSuccess {
		cb.info = info
	} else {
		cb.err = err
	}

	cb.callback <- true
}

func (cb *getInfoCallback) Wait() (string, error) {
	<-cb.callback
	if cb.err == "" {
		return cb.info, nil
	}

	return "", stderrors.New(cb.err)
}
