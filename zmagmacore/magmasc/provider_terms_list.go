package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
)

type (
	TermsList map[string]ProviderTerms
)

var (
	// Make sure ProviderTerms implements Serializable interface.
	_ util.Serializable = (*TermsList)(nil)
)

// Decode implements util.Serializable interface.
func (m *TermsList) Decode(blob []byte) error {
	return json.Unmarshal(blob, m)
}

// Encode implements util.Serializable interface.
func (m *TermsList) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}
