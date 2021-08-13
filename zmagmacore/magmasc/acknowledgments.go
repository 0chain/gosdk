package magmasc

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v3"

	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/storage"
)

type (
	// ActiveAcknowledgments represents active acknowledgments list, used to inserting,
	// removing or getting from state.StateContextI with ActiveAcknowledgmentsKey.
	ActiveAcknowledgments struct {
		Nodes []*Acknowledgment `json:"nodes"`
	}
)

var (
	// Make sure Acknowledgment implements Value interface.
	_ storage.Value = (*ActiveAcknowledgments)(nil)
)

// Encode implements Value interface.
func (m *ActiveAcknowledgments) Encode() []byte {
	val, _ := json.Marshal(m)
	return val
}

// Append appends to stored list of active acknowledgments provided Acknowledgment.
func (m *ActiveAcknowledgments) Append(ackn *Acknowledgment) error {
	const errCode = "append_acknowledgment"

	if _, found := m.GetIndex(ackn.SessionID); !found {
		m.Nodes = append(m.Nodes, ackn)
		if err := storage.GetStorage().SetWithRetries([]byte(ActiveAcknowledgmentsKey), m, 5); err != nil {
			return errors.Wrap(errCode, "error while storing list", err)
		}
	}

	return nil
}

// GetIndex tires to get an acknowledgment form map by given id.
func (m *ActiveAcknowledgments) GetIndex(id string) (int, bool) {
	for idx, item := range m.Nodes {
		if item.SessionID == id {
			return idx, true
		}
	}

	return -1, false
}

// Remove tires to remove an acknowledgment form active list.
func (m *ActiveAcknowledgments) Remove(sessionID string) error {
	const errCode = "append_acknowledgment"

	if sessionID == "" {
		return errors.New(errCode, "acknowledgment is nil")
	}

	if idx, found := m.GetIndex(sessionID); found {
		m.Nodes = append(m.Nodes[:idx], m.Nodes[idx+1:]...)
		if err := storage.GetStorage().SetWithRetries([]byte(ActiveAcknowledgmentsKey), m, 5); err != nil {
			return errors.Wrap(errCode, "error while storing list", err)
		}
	}

	return nil
}

// FetchActiveAcknowledgments extracts active acknowledgments represented in JSON bytes
// stored in storage.Storage with ActiveAcknowledgmentsKey.
//
// If there is no stored active acknowledgments, empty ActiveAcknowledgments will be returned.
func FetchActiveAcknowledgments() (*ActiveAcknowledgments, error) {
	list := &ActiveAcknowledgments{}
	listByt, err := storage.GetStorage().Get([]byte(ActiveAcknowledgmentsKey))
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	} else if err == nil {
		if err := json.Unmarshal(listByt, list); err != nil {
			return nil, err
		}
	}

	return list, nil
}
