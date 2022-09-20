package transaction

import (
	"testing"

	"github.com/0chain/gosdk/core/conf"
	"github.com/stretchr/testify/assert"
)

func TestOptimisticVerificationLearning(t *testing.T) {
	t.Skip()
	conf.InitClientConfig(&conf.Config{
		BlockWorker:             "",
		PreferredBlobbers:       nil,
		MinSubmit:               0,
		MinConfirmation:         50,
		ConfirmationChainLength: 3,
		MaxTxnQuery:             0,
		QuerySleepTime:          0,
		SignatureScheme:         "",
		ChainID:                 "",
		EthereumNode:            "",
	})
	_, err := VerifyTransactionOptimistic("156ba2ee26818513622fcde4b1a7be9e74e81fe5b5370f3451f8e5a12132dfb3",
		[]string{"http://localhost:7171", "http://localhost:7171"})

	assert.NoError(t, err)

}
