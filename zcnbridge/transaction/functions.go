package transaction

// ZCNSC smart contract functions wrappers

import (
	"context"

	"github.com/0chain/gosdk/zcncore"
)

func AddAuthorizer(input *zcncore.AddAuthorizerPayload) (*Transaction, error) {
	t, err := NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	err = t.scheme.ZCNSCAddAuthorizer(input)
	if err != nil {
		return nil, err
	}

	err = t.callBack.waitCompleteCall(context.Background())
	if err != nil {
		return nil, err

	}

	return t, nil
}
