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
	return ExecuteConsumerUpdate(ctx, m)
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
	provider, err := ExecuteProviderRegister(ctx, m)
	if err != nil {
		return nil, errors.Wrap(ErrCodeProviderReg, "error while registering provider", err)
	}
	m.Provider = provider.Provider

	provider, err = ExecuteProviderStake(ctx, m)
	if err != nil {
		return m, nil
	}

	return provider, nil
}

// Update implements registration.Node interface.
func (m *Provider) Update(ctx context.Context) (registration.Node, error) {
	provider, err := ExecuteProviderUpdate(ctx, m)
	if err != nil {
		return nil, errors.Wrap(ErrCodeProviderUpdate, "error while updating provider", err)
	}
	m.Provider = provider.Provider

	provider, err = ExecuteProviderStake(ctx, m)
	if err != nil {
		return m, nil
	}

	return provider, nil
}
