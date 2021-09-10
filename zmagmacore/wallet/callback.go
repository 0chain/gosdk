package wallet

import (
	"github.com/0chain/gosdk/zcncore"
)

type (
	// walletCallback provides callback struct for operations with wallet.
	walletCallback struct{}
)

var (
	// Ensure walletCallback implements interface.
	_ zcncore.WalletCallback = (*walletCallback)(nil)
)

// OnWalletCreateComplete implements zcncore.WalletCallback interface.
func (wb *walletCallback) OnWalletCreateComplete(_ int, _ string, _ string) {}
