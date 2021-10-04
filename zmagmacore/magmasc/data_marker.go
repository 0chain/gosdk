package magmasc

import (
	"encoding/json"

	pb "github.com/0chain/bandwidth_marketplace/code/pb/provider"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zmagmacore/crypto"
	"github.com/0chain/gosdk/zmagmacore/errors"
)

type (
	// DataMarker represents data marker that contains data usage info.
	//
	// Can be interpreted as:
	//
	//  - Regular marker
	//
	//  - QoS marker
	//
	// For more info see DataMarker.IsQoSType implementation
	DataMarker struct {
		*pb.DataMarker
	}
)

var (
	// Make sure User implements Serializable interface.
	_ util.Serializable = (*DataMarker)(nil)
)

// Decode implements util.Serializable interface.
func (m *DataMarker) Decode(blob []byte) error {
	var dataMarker DataMarker
	if err := json.Unmarshal(blob, &dataMarker); err != nil {
		return errDecodeData.Wrap(err)
	}
	if err := dataMarker.Validate(); err != nil {
		return err
	}

	m.DataMarker = &pb.DataMarker{}
	m.UserID = dataMarker.UserID
	m.DataUsage = dataMarker.DataUsage
	m.Qos = dataMarker.Qos
	m.PublicKey = dataMarker.PublicKey
	m.SigScheme = dataMarker.SigScheme
	m.Signature = dataMarker.Signature

	return nil
}

// Encode implements util.Serializable interface.
func (m *DataMarker) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// getSignatureScheme creates zcncrypto.SignatureScheme from data marker's signature scheme type
// and public key.
//
// If data marker's signature scheme is not supported returns error.
func (m *DataMarker) getSignatureScheme() (zcncrypto.SignatureScheme, error) {
	var (
		errCode = "sig_scheme"

		scheme zcncrypto.SignatureScheme
	)
	switch m.SigScheme {
	case "ed25519":
		scheme = zcncrypto.NewED255190chainScheme()

	case "bls0chain":
		scheme = zcncrypto.NewBLS0ChainScheme()

	default:
		return nil, errors.New(errCode, "unsupported signature scheme")
	}

	if err := scheme.SetPublicKey(m.PublicKey); err != nil {
		return nil, err
	}

	return scheme, nil
}

// hashToSign calculates data hash.
func (m *DataMarker) hashToSign() string {
	udm := &DataMarker{
		DataMarker: &pb.DataMarker{
			UserID:    m.UserID,
			DataUsage: m.DataUsage,
			Qos:       m.Qos,
		},
	}
	msg := udm.Encode()
	return crypto.Hash(msg)
}

// IsQoSType checks that DataMarker is QoS type or not.
func (m *DataMarker) IsQoSType() bool {
	return m.UserID != "" || m.PublicKey != "" || m.SigScheme != "" || m.Signature != "" || m.Qos != nil
}

// Sign signs the data.
func (m *DataMarker) Sign(scheme zcncrypto.SignatureScheme, schemeType string) error {
	if len(scheme.GetPrivateKey()) == 0 {
		return errors.New(errCodeBadRequest, "private key is empty")
	}
	sign, err := scheme.Sign(m.hashToSign())
	if err != nil {
		return err
	}

	m.Signature = sign
	m.PublicKey = scheme.GetPublicKey()
	m.SigScheme = schemeType

	return nil
}

// Validate checks the DataMarker for correctness.
// If it is not returns errInvalidDataMarker.
func (m *DataMarker) Validate() (err error) {
	// validate DataUsage, because it is required for all DataMarker's types
	if m.DataUsage == nil || m.DataUsage.SessionID == "" {
		err = errors.New(errCodeBadRequest, "session id is required")
		return errInvalidDataMarker.Wrap(err)
	}

	// if marker is not QoS type, then is no need to continue
	if !m.IsQoSType() {
		return nil
	}

	return m.validateAsQoSType()
}

// validateAsQoSType validates fields of DataMarker which belongs to the QoS marker's type.
func (m *DataMarker) validateAsQoSType() (err error) {
	switch {
	case m.UserID == "":
		err = errors.New(errCodeBadRequest, "user id is required")

	case m.Qos == nil:
		err = errors.New(errCodeBadRequest, "qos is nil")

	case m.Qos.UploadMbps <= 0:
		err = errors.New(errCodeBadRequest, "invalid qos upload mbps")

	case m.Qos.DownloadMbps <= 0:
		err = errors.New(errCodeBadRequest, "invalid qos download mbps")

	case m.Qos.Latency <= 0:
		err = errors.New(errCodeBadRequest, "invalid qos latency")

	case m.PublicKey == "":
		err = errors.New(errCodeBadRequest, "public key is required")

	case m.SigScheme == "":
		err = errors.New(errCodeBadRequest, "signature scheme is required")

	case m.Signature == "":
		err = errors.New(errCodeBadRequest, "signature is required")

	case crypto.Hash(m.PublicKey) != m.UserID:
		err = errors.New(errCodeBadRequest, "user id does not belong to the public key")

	default:
		return nil
	}

	return errInvalidDataMarker.Wrap(err)
}

// Verify verifies data marker's signature.
func (m *DataMarker) Verify() (bool, error) {
	if len(m.PublicKey) == 0 {
		return false, errors.New(errCodeBadRequest, "public key is empty")
	}

	hash := m.hashToSign()

	scheme, err := m.getSignatureScheme()
	if err != nil {
		return false, err
	}

	return scheme.Verify(m.Signature, hash)
}
