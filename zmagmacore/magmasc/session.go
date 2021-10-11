package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/storage"
)

type (
	// Session contains the necessary data obtained when the consumer
	// accepts the provider terms and stores in the state of the blockchain
	// as a result of performing the consumerAcceptTerms MagmaSmartContract function.
	Session struct {
		SessionID   string       `json:"session_id"`
		Billing     Billing      `json:"billing"`
		Consumer    *Consumer    `json:"consumer,omitempty"`
		Provider    *Provider    `json:"provider,omitempty"`
		AccessPoint *AccessPoint `json:"access_point,omitempty"`
		TokenPool   *TokenPool   `json:"token_pool,omitempty"`
	}
)

var (
	// Make sure Session implements PoolConfigurator interface.
	_ PoolConfigurator = (*Session)(nil)

	// Make sure Session implements Value interface.
	_ storage.Value = (*Session)(nil)

	// Make sure Session implements Serializable interface.
	_ util.Serializable = (*Session)(nil)
)

// ActiveKey returns key used for operations with storage.Storage
// SessionPrefix + SessionActivePrefixPart + Session.SessionID.
func (s *Session) ActiveKey() []byte {
	return []byte(SessionPrefix + SessionActivePrefixPart + s.SessionID)
}

// Decode implements util.Serializable interface.
func (s *Session) Decode(blob []byte) error {
	var sess Session
	if err := json.Unmarshal(blob, &sess); err != nil {
		return ErrDecodeData.Wrap(err)
	}
	if err := sess.Validate(); err != nil {
		return err
	}

	s.SessionID = sess.SessionID
	s.Billing = sess.Billing
	s.Consumer = sess.Consumer
	s.Provider = sess.Provider
	s.AccessPoint = sess.AccessPoint
	s.TokenPool = sess.TokenPool

	return nil
}

// Encode implements util.Serializable interface.
func (s *Session) Encode() []byte {
	blob, _ := json.Marshal(s)
	return blob
}

// Key returns key with SessionPrefix.
// Used for operations with storage.Storage.
func (s *Session) Key() []byte {
	return []byte(SessionPrefix + s.SessionID)
}

// PoolBalance implements PoolConfigurator interface.
func (s *Session) PoolBalance() int64 {
	return s.AccessPoint.Terms.GetAmount()
}

// PoolID implements PoolConfigurator interface.
func (s *Session) PoolID() string {
	return s.SessionID
}

// PoolHolderID implements PoolConfigurator interface.
func (s *Session) PoolHolderID() string {
	return Address
}

// PoolPayerID implements PoolConfigurator interface.
func (s *Session) PoolPayerID() string {
	return s.Consumer.ID
}

// PoolPayeeID implements PoolConfigurator interface.
func (s *Session) PoolPayeeID() string {
	return s.AccessPoint.ID
}

// Validate checks Session for correctness.
// If it is not return errInvalidSession.
func (s *Session) Validate() (err error) {
	switch { // is invalid
	case s.SessionID == "":
		err = errors.New(ErrCodeBadRequest, "session id is required")

	case s.AccessPoint == nil || s.AccessPoint.ID == "":
		err = errors.New(ErrCodeBadRequest, "access point id is required")

	case s.Consumer == nil || s.Consumer.ExtID == "":
		err = errors.New(ErrCodeBadRequest, "consumer external id is required")

	case s.Provider == nil || s.Provider.ExtID == "":
		err = errors.New(ErrCodeBadRequest, "provider external id is required")

	default:
		return nil // is valid
	}

	return ErrInvalidSession.Wrap(err)
}
