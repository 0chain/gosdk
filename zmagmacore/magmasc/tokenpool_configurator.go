package magmasc

type (
	// PoolConfigurator represents a pool config interface.
	PoolConfigurator interface {
		// PoolBalance returns the amount value of token pool.
		PoolBalance() int64

		// PoolID returns the token pool ID.
		PoolID() string

		// PoolHolderID returns the token pool holder ID.
		PoolHolderID() string

		// PoolPayerID returns the token pool payer ID.
		PoolPayerID() string

		// PoolPayeeID returns the token pool payee ID.
		PoolPayeeID() string
	}
)
