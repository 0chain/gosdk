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
	ov := NewOptimisticVerifier([]string{"https://dev.zus.network/sharder01", "https://dev.zus.network/sharder02"})
	_, err := ov.VerifyTransactionOptimistic("a20360964c067b319d52b5cad71d771b0e1d2a80e76001da73009899b09ffa31")

	assert.NoError(t, err)

}
