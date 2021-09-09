package registration

import (
	"context"
)

type (
	// Node represent bandwidth-marketplace node that can be registered and updated.
	Node interface {
		// Register executes registration function of the Magma Smart Contract.
		Register(context.Context) (Node, error)

		// Update executes update function of the Magma Smart Contract.
		Update(context.Context) (Node, error)

		// IsNodeRegistered calls rest point of the Magma Smart Contract and checks is node already registered.
		IsNodeRegistered() (bool, error)
	}
)
