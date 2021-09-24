package magmasc

import (
	"encoding/json"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zmagmacore/config"
	"github.com/0chain/gosdk/zmagmacore/crypto"
	"github.com/0chain/gosdk/zmagmacore/time"
)

type (
	// UserDataMarker represents user stored in blockchain.
	UserDataMarker struct {
		UserID     string         `json:"user_id"`
		ProviderID string         `json:"provider_id"`
		SessionID  string         `json:"session_id"`
		DataUsage  DataUsage      `json:"data_usage"`
		QoS        QoS            `json:"qos"`
		Timestamp  time.Timestamp `json:"timestamp,omitempty"`
		Signature  string         `json:"signature,omitempty"`
	}

	// QoS represents config of qos.
	QoS struct {
		DownloadMbps float32 `json:"download_mbps"`
		UploadMbps   float32 `json:"upload_mbps"`
		Latency      float32 `json:"latency"`
	}
)

var (
	// Make sure User implements Serializable interface.
	_ util.Serializable = (*UserDataMarker)(nil)
)

// NewUserDataMarkerFromCfg creates UserDataMarker from config.UserDataMarker.
func NewUserDataMarkerFromCfg(cfg config.UserDataMarker) *UserDataMarker {
	return &UserDataMarker{
		UserID:     cfg.UserID,
		ProviderID: cfg.ProviderID,
		SessionID:  cfg.SessionID,
		DataUsage: DataUsage{
			DownloadBytes: cfg.DataUsage.DownloadBytes,
			UploadBytes:   cfg.DataUsage.UploadBytes,
			SessionID:     cfg.DataUsage.SessionID,
			SessionTime:   cfg.DataUsage.SessionTime},
		QoS: QoS{
			DownloadMbps: cfg.Qos.DownloadMbps,
			UploadMbps:   cfg.Qos.UploadMbps,
			Latency:      cfg.Qos.Latency,
		},
	}
}

// Decode implements util.Serializable interface.
func (m *UserDataMarker) Decode(blob []byte) error {
	var dataMarker UserDataMarker
	if err := json.Unmarshal(blob, &dataMarker); err != nil {
		return errDecodeData.Wrap(err)
	}
	if err := dataMarker.Validate(); err != nil {
		return err
	}

	m.UserID = dataMarker.UserID
	m.ProviderID = dataMarker.ProviderID
	m.SessionID = dataMarker.SessionID
	m.DataUsage = dataMarker.DataUsage
	m.QoS = dataMarker.QoS
	m.Timestamp = dataMarker.Timestamp
	m.Signature = dataMarker.Signature

	return nil
}

// Encode implements util.Serializable interface.
func (m *UserDataMarker) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// Validate checks the UserDataMarker for correctness.
// If it is not return errInvalidUserDataMarker.
func (m *UserDataMarker) Validate() (err error) {
	switch { // is invalid
	case m.UserID == "":
		err = errors.New(errCodeBadRequest, "user data marker external id is required")

	default:
		return nil // is valid
	}

	return errInvalidUserDataMarker.Wrap(err)
}

// Sign signs the data.
func (m *UserDataMarker) Sign(scheme zcncrypto.SignatureScheme) error {
	if len(scheme.GetPrivateKey()) == 0 {
		return errors.New(errCodeBadRequest, "private key does not exists for signing")
	}
	sign, err := scheme.Sign(m.hashToSign())
	if err != nil {
		return err
	}

	m.Signature = sign

	return nil
}

// Verify makes a check signature verification.
func (m *UserDataMarker) Verify(scheme zcncrypto.SignatureScheme) (bool, error) {
	if len(scheme.GetPublicKey()) == 0 {
		return false, errors.New(errCodeBadRequest, "public key does not exists for verification")
	}
	hash := m.hashToSign()

	return scheme.Verify(m.Signature, hash)
}

// hashToSign calculates data hash.
func (m *UserDataMarker) hashToSign() string {
	udm := *m
	udm.Signature = ""

	return crypto.Hash(udm.Encode())
}
