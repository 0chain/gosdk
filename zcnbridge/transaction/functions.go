package transaction

// ZCNSC smart contract functions wrappers

import (
	"context"

	"github.com/0chain/gosdk/zcncore"
)

// AddAuthorizer adds authorizer to the bridge
//   - ctx is the context of the request.
//   - input is the payload of the request.
func AddAuthorizer(ctx context.Context, input *zcncore.AddAuthorizerPayload) (Transaction, error) {
	t, err := NewTransactionEntity(0)
	if err != nil {
		return nil, err
	}

	scheme := t.GetScheme()

	err = scheme.ZCNSCAddAuthorizer(input)
	if err != nil {
		return t, err
	}

	callBack := t.GetCallback()

	err = callBack.WaitCompleteCall(ctx)
	t.SetHash(scheme.Hash())
	if err != nil {
		return t, err
	}

	return t, nil
}

// AuthorizerHealthCheck performs health check of the authorizer
//   - ctx is the context of the request.
//   - input is the payload of the request.
func AuthorizerHealthCheck(ctx context.Context, input *zcncore.AuthorizerHealthCheckPayload) (Transaction, error) {
	t, err := NewTransactionEntity(0)
	if err != nil {
		return nil, err
	}

	scheme := t.GetScheme()

	err = scheme.ZCNSCAuthorizerHealthCheck(input)
	if err != nil {
		return t, err
	}

	callBack := t.GetCallback()

	err = callBack.WaitCompleteCall(ctx)
	t.SetHash(scheme.Hash())
	if err != nil {
		return t, err
	}

	return t, nil
}
