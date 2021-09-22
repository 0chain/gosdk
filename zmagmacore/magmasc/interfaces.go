package magmasc

import (
	"context"

	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/registration"
)

var (
	// Ensure Consumer implements registration.Node interface.
	_ registration.Node = (*Consumer)(nil)
)

// IsNodeRegistered implements registration.Node interface.
func (m *Consumer) IsNodeRegistered() (bool, error) {
	return IsConsumerRegisteredRP(m.ExtID)
}

// Register implements registration.Node interface.
func (m *Consumer) Register(ctx context.Context) (registration.Node, error) {
	return ExecuteConsumerRegister(ctx, m)
}

// Update implements registration.Node interface.
func (m *Consumer) Update(ctx context.Context) (registration.Node, error) {
	return ExecuteConsumerRegister(ctx, m)
}

var (
	// Ensure Provider implements registration.Node interface.
	_ registration.Node = (*Provider)(nil)
)

// IsNodeRegistered implements registration.Node interface.
func (m *Provider) IsNodeRegistered() (bool, error) {
	return IsProviderRegisteredRP(m.ExtID)
}

// Register implements registration.Node interface.
func (m *Provider) Register(ctx context.Context) (registration.Node, error) {
	var (
		errCode = "provider_register"

		minStake int64
	)
	if m.minStake {
		var err error
		minStake, err = ProviderMinStakeFetch()
		if err != nil {
			return nil, errors.Wrap(errCode, "error while fetching min stake", err)
		}
	}

	m.MinStake = minStake
	return ExecuteProviderRegister(ctx, m)
}

// Update implements registration.Node interface.
func (m *Provider) Update(ctx context.Context) (registration.Node, error) {
	var (
		errCode = "provider_update"

		minStake int64
	)
	if m.minStake {
		var err error
		minStake, err = ProviderMinStakeFetch()
		if err != nil {
			return nil, errors.Wrap(errCode, "error while fetching min stake", err)
		}
	}

	m.MinStake = minStake

	return ExecuteProviderUpdate(ctx, m)
}
