package magmasc

const (
	// Address represents the address of the Magma smart contract.
	// Used while making requests to smart contract's rest points and executing smart contracts functions.
	Address = "11f8411db41e34cea7c100f19faff32da8f3cd5a80635731cec06f32d08089be"

	// Name contents the smart contract name.
	Name = "magma"

	// Colon represents values separator.
	Colon = ":"

	// ProviderMinStake represents the key of a provider min stake config.
	ProviderMinStake = "provider.min_stake"

	// ServiceCharge represents the key of a service charge config.
	ServiceCharge = "service_charge"

	// RootPath describes the magma smart contract's root path.
	RootPath = ".0chain.net"

	// StoreName describes the magma smart contract's store name.
	StoreName = "magmadb"

	// StorePath describes the magma smart contract's store path.
	StorePath = "data/rocksdb/magmasc"

	// GetAllConsumersRP represents MagmaSC relative path.
	// Used to list all registered in the blockchain consumers.
	GetAllConsumersRP = "/allConsumers"

	// GetAllProvidersRP represents MagmaSC relative path.
	// Used to list all registered in the blockchain providers.
	GetAllProvidersRP = "/allProviders"
)

const (
	// AllRewardPoolsKey is a concatenated Address
	// and SHA3-256 hex encoded hash of "all_reward_pools" string.
	AllRewardPoolsKey = Address + "59864241d642b4b6b5e5998b70bd201ca4d48926de8934e02e300950c778c7c2"

	// RewardPoolLock represents the name of MagmaSmartContract function.
	// When function is called it means that wallet creates a new locked token pool.
	RewardPoolLock = "reward_pool_lock"

	// RewardPoolUnlock represents the name of MagmaSmartContract function.
	// When function is called it means that wallet refunds a locked token pool.
	RewardPoolUnlock = "reward_pool_unlock"

	// RewardTokenPool contents a value of reward token pool string type.
	RewardTokenPool = "reward_token_pool"
)

const (
	// SessionPrefix represents prefix to save into storage.
	SessionPrefix = "session"

	// SessionActivePrefixPart represents prefix part to save into storage.
	SessionActivePrefixPart = "act"

	// SessionRP represents MagmaSC relative path.
	// Used to retrieve accepted session.
	SessionRP = "/sessionAccepted"

	// IsSessionExistRP represents MagmaSC relative path.
	// Used to check existing of session.
	IsSessionExistRP = "/sessionExist"

	// VerifySessionAcceptedRP represents MagmaSC relative path.
	// Used to verify accepting Provider's terms by Consumer.
	VerifySessionAcceptedRP = "/sessionAcceptedVerify"
)

const (
	// AllConsumersKey is a concatenated Address
	// and SHA3-256 hex encoded hash of "all_consumers" string.
	AllConsumersKey = Address + "226fe0dc53026203416c348f675ce0c5ea35d87d959e41aaf6a3ca7829741710"

	// ConsumerRegisterFuncName represents MagmaSC function.
	// Used to register bandwidth-marketplace's node.
	ConsumerRegisterFuncName = "consumer_register"

	// ConsumerSessionStartFuncName represents MagmaSC function.
	// Used to start session.
	ConsumerSessionStartFuncName = "consumer_session_start"

	// ConsumerSessionStopFuncName represents MagmaSC function.
	// Used to stop session.
	ConsumerSessionStopFuncName = "consumer_session_stop"

	// ConsumerUpdateFuncName represents MagmaSC function.
	// Used to update consumer node info.
	ConsumerUpdateFuncName = "consumer_update"

	// ConsumerFetchRP represents MagmaSC relative path.
	// Used to fetch consumer info.
	ConsumerFetchRP = "/consumerFetch"

	// ConsumerRegisteredRP represents MagmaSC relative path.
	// Used to fetch consumer registered info.
	ConsumerRegisteredRP = "/consumerExist"

	// ConsumerType contents a value of consumer node type.
	ConsumerType = "consumer"
)

const (
	// TermsExpiredDuration represents value for
	// minimal duration of provider terms that will pass check it's expired.
	TermsExpiredDuration = 1 * 60 // 1 minute

	// AllProvidersKey is a concatenated Address
	// and SHA3-256 hex encoded hash of "all_providers" string.
	AllProvidersKey = Address + "7e306c02ea1719b598aaf9dc7516eb930cd47c5360d974e22ab01e21d66a93d8"

	// ProviderDataUsageFuncName represents MagmaSC function.
	// Used to update session info about data usages and collecting payments data
	// from consumer to provider.
	ProviderDataUsageFuncName = "provider_data_usage"

	// ProviderRegisterFuncName represents MagmaSC function.
	// Used to register bandwidth-marketplace's node.
	ProviderRegisterFuncName = "provider_register"

	// ProviderUpdateFuncName represents MagmaSC function.
	// Used for updating provider terms.
	ProviderUpdateFuncName = "provider_update"

	// ProviderStakeFuncName represents MagmaSC function.
	// Used to stake tokens to MagmaSC.
	ProviderStakeFuncName = "provider_stake"

	// ProviderUnStakeFuncName represents MagmaSC function.
	// Used for unstake tokens to MagmaSC.
	ProviderUnStakeFuncName = "provider_unstake"

	// ProviderFetchRP represents MagmaSC relative path.
	// Used to fetch provider info.
	ProviderFetchRP = "/providerFetch"

	// ProviderMinStakeFetchRP represents MagmaSC relative path
	// Used to fetch provider min stake
	ProviderMinStakeFetchRP = "/providerMinStakeFetch"

	// ProviderRegisteredRP represents MagmaSC relative path.
	// Used to fetch provider registered info.
	ProviderRegisteredRP = "/providerExist"

	// ProviderType contents a value of provider node type.
	ProviderType = "provider"

	// ProviderStakeTokenPool contents a value of provider's stake token pool string type.
	ProviderStakeTokenPool = "provider_stake_token_pool"
)

const (
	// AllAccessPointsKey is a concatenated Address
	// and SHA3-256 hex encoded hash of "all_access_points" string.
	AllAccessPointsKey = Address // + "7e306c02ea1719b598aaf9dc7516eb930cd47c5360d974e22ab01e21d66a93d8" TODO

	// AccessPointRegisterFuncName represents MagmaSC function.
	// Used to register bandwidth-marketplace's node.
	AccessPointRegisterFuncName = "access_point_register"

	// AccessPointUpdateFuncName represents MagmaSC function.
	// Used to update access point node info.
	AccessPointUpdateFuncName = "access_point_update"

	// AccessPointFetchRP represents MagmaSC relative path.
	// Used to fetch access point info.
	AccessPointFetchRP = "/accessPointFetch"

	// AccessPointMinStakeFetchRP represents MagmaSC relative path.
	// Used to fetch min stake configured value.
	AccessPointMinStakeFetchRP = "/accessPointMinStakeFetch"

	// AccessPointRegisteredRP represents MagmaSC relative path.
	// Used to fetch access point registered info.
	AccessPointRegisteredRP = "/accessPointExist"

	// AccessPointType contents a value of type of Access Point's node.
	AccessPointType = "access-point"

	// AccessPointMinStake represents the key of a access point min stake config.
	AccessPointMinStake = "access_point.min_stake"

	// AccessPointStakeTokenPool contents a value of access point's stake token pool string type.
	AccessPointStakeTokenPool = "access_point_stake_token_pool"
)

const (
	// FetchBillingRatioRP represents MagmaSC relative path.
	// Used to fetch billing ratio.
	FetchBillingRatioRP = "/fetchBillingRatio"
)

const (
	// one billion (Giga) is a unit prefix in metric systems
	// of units denoting a factor of one billion (1e9 or 1_000_000_000).
	billion = 1e9

	// one million (Mega) is a unit prefix in metric systems
	// of units denoting a factor of one million (1e6 or 1_000_000).
	million = 1e6

	// octet represents number of bits in an octet.
	octet = 8
)

const (
	// UserRegisterFuncName represents name for User's registration MagmaSmartContract function.
	UserRegisterFuncName = "user_register"

	// UserUpdateFuncName represents name for user data update MagmaSmartContract function.
	UserUpdateFuncName = "user_update"

	// UserFetchRP represents MagmaSC relative path.
	// Used to fetch user info.
	UserFetchRP = "/userFetch"

	// UserRegisteredRP represents MagmaSC relative path.
	// Used to fetch user registered info.
	UserRegisteredRP = "/userExist"
)
