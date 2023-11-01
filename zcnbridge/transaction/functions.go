package transaction

// ZCNSC smart contract functions wrappers

import (
	"context"

	"github.com/0chain/gosdk/zcncore"
)

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
