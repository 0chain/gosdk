package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zmagmacore/config"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/node"
)

type (
	// Consumer represents consumers node stored in blockchain.
	Consumer struct {
		ID    string `json:"id"`
		ExtID string `json:"ext_id"`
		Host  string `json:"host,omitempty"`
	}
)

var (
	// Make sure Consumer implements Serializable interface.
	_ util.Serializable = (*Consumer)(nil)
)

// NewConsumerFromCfg creates Consumer from config.Consumer.
func NewConsumerFromCfg(cfg *config.Consumer) *Consumer {
	return &Consumer{
		ID:    node.ID(),
		ExtID: cfg.ExtID,
		Host:  cfg.Host,
	}
}

// Decode implements util.Serializable interface.
func (m *Consumer) Decode(blob []byte) error {
	var consumer Consumer
	if err := json.Unmarshal(blob, &consumer); err != nil {
		return errDecodeData.Wrap(err)
	}
	if err := consumer.Validate(); err != nil {
		return err
	}

	m.ID = consumer.ID
	m.ExtID = consumer.ExtID
	m.Host = consumer.Host

	return nil
}

// Encode implements util.Serializable interface.
func (m *Consumer) Encode() []byte {
	blob, _ := json.Marshal(m)
	return blob
}

// GetType returns node type.
func (m *Consumer) GetType() string {
	return consumerType
}

// Validate checks the Consumer for correctness.
// If it is not return errInvalidConsumer.
func (m *Consumer) Validate() (err error) {
	switch { // is invalid
	case m.ExtID == "":
		err = errors.New(errCodeBadRequest, "consumer external id is required")

	default:
		return nil // is valid
	}

	return errInvalidConsumer.Wrap(err)
}
