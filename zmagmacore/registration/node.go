package registration

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/0chain/gosdk/zmagmacore/errors"
	"github.com/0chain/gosdk/zmagmacore/log"
)

// RegisterOrUpdateWithRetries registers bandwidth-marketplace Node in blockchain
// if node is not registered or updates if is with retries.
//
// If an error occurs during execution, the program terminates with code 2 and the last error will be written in os.Stderr.
//
// RegisterOrUpdateWithRetries should be used only once while application is starting.
func RegisterOrUpdateWithRetries(ctx context.Context, bmNode Node, numTries int) {
	var (
		executeSC func(ctx context.Context) (Node, error)
	)

	registered, err := bmNode.IsNodeRegistered()
	if err != nil {
		errors.ExitErr("err while checking nodes registration", err, 2)
	}
	if registered {
		executeSC = bmNode.Update
		log.Logger.Debug("Node is already registered in the blockchain, trying to update ...")
	} else {
		executeSC = bmNode.Register
		log.Logger.Debug("Node is not registered in the Blockchain, trying to start registration ...")
	}

	for ind := 0; ind < numTries; ind++ {
		var respNode Node
		respNode, err = executeSC(ctx)
		if err != nil {
			log.Logger.Debug("Executing smart contract failed. Sleep for 1 seconds ...",
				zap.String("err", err.Error()),
			)
			time.Sleep(time.Second)
			continue
		}

		log.Logger.Info("Node is registered in the blockchain", zap.Any("node", respNode))
		break
	}

	if err != nil {
		errors.ExitErr("error while registering", err, 2)
	}
}
