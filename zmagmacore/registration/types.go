package registration

import (
	"context"

	"github.com/0chain/gosdk/zmagmacore/magmasc"
)

type (
	// Node represent bandwidth-marketplace node that can be registered.
	Node interface {
		// RegistrationFuncName returns name of magma sc function used for registration Node.
		RegistrationFuncName() string

		// UpdateNodeFuncName returns name of magma sc function used for updating Node.
		UpdateNodeFuncName() string

		// IsNodeRegisteredRP returns name of magma sc relative path for checking registration of Node.
		IsNodeRegisteredRP() string

		// ExternalID returns external ID of Node
		ExternalID() string

		// Encode encodes Node in json bytes.
		Encode() []byte
	}

	// executeSmartContract represent types for functions that executes sc functions.
	executeSmartContract func(ctx context.Context, node Node) (string, error)
)

var (
	// Ensure Consumer implements interface.
	_ Node = (*magmasc.Consumer)(nil)

	// Ensure Provider implements interface.
	_ Node = (*magmasc.Provider)(nil)
)
