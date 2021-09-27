package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/errors"
)

type (
	// DataUsage represents session data sage implementation.
	DataUsage struct {
		DownloadBytes uint64 `json:"download_bytes"`
		UploadBytes   uint64 `json:"upload_bytes"`
		SessionID     string `json:"session_id"`
		SessionTime   uint32 `json:"session_time"`
	}
)

var (
	// Make sure DataUsage implements Serializable interface.
	_ util.Serializable = (*DataUsage)(nil)
)

// Decode implements util.Serializable interface.
func (m *DataUsage) Decode(blob []byte) error {
	var dataUsage DataUsage
	if err := json.Unmarshal(blob, &dataUsage); err != nil {
		return errDecodeData.Wrap(err)
	}
	if err := dataUsage.Validate(); err != nil {
		return err
	}

	m.DownloadBytes = dataUsage.DownloadBytes
	m.UploadBytes = dataUsage.UploadBytes
	m.SessionID = dataUsage.SessionID
	m.SessionTime = dataUsage.SessionTime

	return nil
}

// Encode implements util.Serializable interface.
func (m *DataUsage) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// Validate checks DataUsage for correctness.
func (m *DataUsage) Validate() (err error) {
	switch { // is invalid
	case m.SessionID == "":
		err = errors.New(errCodeBadRequest, "session id is required")

	default: // is valid
		return nil
	}

	return errInvalidDataUsage.Wrap(err)
}
