package magmasc

import (
	"encoding/json"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/config"
)

// User represent user in blockchain.
type User struct {
	ID         string `json:"id,omitempty"`
	ConsumerID string `json:"consumer_id,omitempty"`
}

var (
	// Make sure User implements Serializable interface.
	_ util.Serializable = (*User)(nil)
)

// NewUserFromCfg creates User from config.User
func NewUserFromCfg(cfg *config.User) *User {
	return &User{
		ID:         cfg.ID,
		ConsumerID: cfg.ConsumerID,
	}
}

// Decode implements util.Serializable interface.
func (m *User) Decode(blob []byte) error {
	var user User
	if err := json.Unmarshal(blob, &user); err != nil {
		return errDecodeData.Wrap(err)
	}
	if err := user.Validate(); err != nil {
		return err
	}

	m.ID = user.ID
	m.ConsumerID = user.ConsumerID

	return nil
}

// Encode implements util.Serializable interface.
func (m *User) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// Validate checks the User for correctness.
// If it is not return errInvalidUser.
func (m *User) Validate() (err error) {
	switch { // is invalid
	case m.ID == "":
		err = errors.New(errCodeBadRequest, "user id is required")

	case m.ConsumerID == "":
		err = errors.New(errCodeBadRequest, "user consumer id is required")

	default:
		return nil // is valid
	}

	return errInvalidUser.Wrap(err)
}
