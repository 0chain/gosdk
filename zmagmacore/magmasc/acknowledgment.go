// DEPRECATED: This package is deprecated and will be removed in a future release.
package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/storage"
)

type (
	// Acknowledgment contains the necessary data obtained when the consumer
	// accepts the provider terms and stores in the state of the blockchain
	// as a result of performing the consumerAcceptTerms MagmaSmartContract function.
	Acknowledgment struct {
		SessionID     string        `json:"session_id"`
		AccessPointID string        `json:"access_point_id"`
		Billing       Billing       `json:"billing"`
		Consumer      *Consumer     `json:"consumer,omitempty"`
		Provider      *Provider     `json:"provider,omitempty"`
		Terms         ProviderTerms `json:"terms"`
		TokenPool     *TokenPool    `json:"token_pool,omitempty"`
	}
)

var (
	// Make sure Acknowledgment implements PoolConfigurator interface.
	_ PoolConfigurator = (*Acknowledgment)(nil)

	// Make sure Acknowledgment implements Value interface.
	_ storage.Value = (*Acknowledgment)(nil)

	// Make sure Acknowledgment implements Serializable interface.
	_ util.Serializable = (*Acknowledgment)(nil)
)

// ActiveKey returns key used for operations with storage.Storage
// AcknowledgmentPrefix + AcknowledgmentActivePrefixPart + Acknowledgment.SessionID.
func (m *Acknowledgment) ActiveKey() []byte {
	return []byte(AcknowledgmentPrefix + AcknowledgmentActivePrefixPart + m.SessionID)
}

// Decode implements util.Serializable interface.
func (m *Acknowledgment) Decode(blob []byte) error {
	var ackn Acknowledgment
	if err := json.Unmarshal(blob, &ackn); err != nil {
		return errDecodeData.Wrap(err)
	}
	if err := ackn.Validate(); err != nil {
		return err
	}

	m.SessionID = ackn.SessionID
	m.AccessPointID = ackn.AccessPointID
	m.Billing = ackn.Billing
	m.Consumer = ackn.Consumer
	m.Provider = ackn.Provider
	m.Terms = ackn.Terms
	m.TokenPool = ackn.TokenPool

	return nil
}

// Encode implements util.Serializable interface.
func (m *Acknowledgment) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// Key returns key with AcknowledgmentPrefix.
// Used for operations with storage.Storage.
func (m *Acknowledgment) Key() []byte {
	return []byte(AcknowledgmentPrefix + m.SessionID)
}

// PoolBalance implements PoolConfigurator interface.
func (m *Acknowledgment) PoolBalance() uint64 {
	return m.Terms.GetAmount()
}

// PoolID implements PoolConfigurator interface.
func (m *Acknowledgment) PoolID() string {
	return m.SessionID
}

// PoolHolderID implements PoolConfigurator interface.
func (m *Acknowledgment) PoolHolderID() string {
	return Address
}

// PoolPayerID implements PoolConfigurator interface.
func (m *Acknowledgment) PoolPayerID() string {
	return m.Consumer.ID
}

// PoolPayeeID implements PoolConfigurator interface.
func (m *Acknowledgment) PoolPayeeID() string {
	return m.Provider.ID
}

// Validate checks Acknowledgment for correctness.
// If it is not return errInvalidAcknowledgment.
func (m *Acknowledgment) Validate() (err error) {
	switch { // is invalid
	case m.SessionID == "":
		err = errors.New(errCodeBadRequest, "session id is required")

	case m.AccessPointID == "":
		err = errors.New(errCodeBadRequest, "access point id is required")

	case m.Consumer == nil || m.Consumer.ExtID == "":
		err = errors.New(errCodeBadRequest, "consumer external id is required")

	case m.Provider == nil || m.Provider.ExtID == "":
		err = errors.New(errCodeBadRequest, "provider external id is required")

	default:
		return nil // is valid
	}

	return errInvalidAcknowledgment.Wrap(err)
}
