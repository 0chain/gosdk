// DEPRECATED: This package is deprecated and will be removed in a future release.
package registration

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/http"
	"github.com/0chain/gosdk/zmagmacore/log"
	"github.com/0chain/gosdk/zmagmacore/magmasc"
	"github.com/0chain/gosdk/zmagmacore/transaction"
)

// RegisterOrUpdateWithRetries registers bandwidth-marketplace Node in blockchain
// if node is not registered or updates if is with retries.
//
// If an error occurs during execution, the program terminates with code 2 and the last error will be written in os.Stderr.
//
// RegisterOrUpdateWithRetries should be used only once while application is starting.
func RegisterOrUpdateWithRetries(ctx context.Context, bmNode Node, numTries int) {
	var (
		executeSC executeSmartContract
	)

	registered, err := isNodeRegistered(bmNode)
	if err != nil {
		errors.ExitErr("err while checking nodes registration", err, 2)
	}
	if registered {
		executeSC = update
		log.Logger.Debug("Node is already registered in the blockchain, trying to update ...")
	} else {
		executeSC = register
		log.Logger.Debug("Node is not registered in the Blockchain, trying to start registration ...")
	}

	var (
		txnOutput string
	)
	for ind := 0; ind < numTries; ind++ {
		txnOutput, err = executeSC(ctx, bmNode)
		if err != nil {
			log.Logger.Debug("Executing smart contract failed. Sleep for 1 seconds ...",
				zap.String("err", err.Error()),
			)
			sys.Sleep(time.Second)
			continue
		}

		log.Logger.Info("Node is registered in the blockchain", zap.Any("txn_output", txnOutput))
		break
	}

	if err != nil {
		errors.ExitErr("error while registering", err, 2)
	}
}

func isNodeRegistered(bmNode Node) (bool, error) {
	params := map[string]string{
		"ext_id": bmNode.ExternalID(),
	}
	registeredByt, err := http.MakeSCRestAPICall(magmasc.Address, bmNode.IsNodeRegisteredRP(), params)
	if err != nil {
		return false, err
	}

	var registered bool
	if err := json.Unmarshal(registeredByt, &registered); err != nil {
		return false, err
	}

	return registered, nil
}

func register(ctx context.Context, bmNode Node) (txnOutput string, err error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return "", err
	}

	input := bmNode.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, magmasc.Address, bmNode.RegistrationFuncName(), string(input), 0)
	if err != nil {
		return "", err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return "", err
	}

	return txn.TransactionOutput, nil
}

func update(ctx context.Context, bmNode Node) (txnOutput string, err error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return "", err
	}

	input := bmNode.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, magmasc.Address, bmNode.UpdateNodeFuncName(), string(input), 0)
	if err != nil {
		return "", err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return "", err
	}

	return txn.TransactionOutput, nil
}
