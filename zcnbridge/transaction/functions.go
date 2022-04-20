package transaction

// ZCNSC smart contract functions wrappers. Partially covered. At this stage not all required

import (
	"context"
	"encoding/json"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zcncore"
)

func (t *Transaction) AddAuthorizer(ctx context.Context, input *zcncore.AddAuthorizerPayload) error {
	t, err := NewTransactionEntity()
	if err != nil {
		return err
	}

	buffer, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = t.ExecuteSmartContract(
		ctx,
		zcncore.ZCNSCSmartContractAddress,
		transaction.ZCNSC_ADD_AUTHORIZER,
		string(buffer),
		0,
	)
	if err != nil {
		return err
	}

	return nil
}
