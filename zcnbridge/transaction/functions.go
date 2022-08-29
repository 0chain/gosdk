package transaction

// ZCNSC smart contract functions wrappers

import (
	"context"

	"github.com/0chain/gosdk/zcncore"
)

func AddAuthorizer(ctx context.Context, input *zcncore.AddAuthorizerPayload) (*Transaction, error) {
	t, err := NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	err = t.scheme.ZCNSCAddAuthorizer(input)
	if err != nil {
		return t, err
	}

	err = t.callBack.waitCompleteCall(ctx)
	t.Hash = t.scheme.Hash()
	if err != nil {
		return t, err
	}

	return t, nil
}

func DeleteAuthorizer(ctx context.Context, input *zcncore.DeleteAuthorizerPayload) (*Transaction, error) {
	t, err := NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	err = t.scheme.ZCNSCDeleteAuthorizer(input)
	if err != nil {
		return t, err
	}

	err = t.callBack.waitCompleteCall(ctx)
	t.Hash = t.scheme.Hash()
	if err != nil {
		return t, err
	}

	return t, nil
}
