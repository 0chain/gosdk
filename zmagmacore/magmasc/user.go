package magmasc

import (
	"encoding/json"
	"os"

	"github.com/0chain/errors"
	"gopkg.in/yaml.v3"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/magmasc/pb"
)

type (
	// User wraps proto.user for blockchain use.
	User struct {
		// pb.User embedded struct for implement our interface.
		*pb.User
	}
)

var (
	// Make sure User implements Serializable interface.
	_ util.Serializable = (*User)(nil)
)

// Decode implements util.Serializable interface.
func (m *User) Decode(blob []byte) error {
	var user User
	if err := json.Unmarshal(blob, &user); err != nil {
		return ErrDecodeData.Wrap(err)
	}
	if err := user.Validate(); err != nil {
		return err
	}

	m.User = user.User

	return nil
}

// Encode implements util.Serializable interface.
func (m *User) Encode() []byte {
	blob, _ := json.Marshal(m.User)
	return blob
}

// Validate checks the User for correctness.
// If it is not return errInvalidUser.
func (m *User) Validate() (err error) {
	switch { // is invalid
	case m.User == nil:
		err = errors.New(errCodeBadRequest, "user is not present yet")

	case m.ID == "":
		err = errors.New(ErrCodeBadRequest, "user id is required")

	case m.ConsumerID == "":
		err = errors.New(ErrCodeBadRequest, "user consumer id is required")

	default:
		return nil // is valid
	}

	return ErrInvalidUser.Wrap(err)
}

// ReadYAML reads config yaml file from path.
func (m *User) ReadYAML(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) { _ = f.Close() }(f)

	decoder := yaml.NewDecoder(f)

	return decoder.Decode(m)
}
