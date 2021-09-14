package magmasc

const (
	// Address represents the address of the Magma smart contract.
	// Used while making requests to smart contract's rest points and executing smart contracts functions.
	Address = "11f8411db41e34cea7c100f19faff32da8f3cd5a80635731cec06f32d08089be"

	// GetAllConsumersRP represents MagmaSC relative path.
	// Used to list all registered in the blockchain consumers.
	GetAllConsumersRP = "/allConsumers"

	// GetAllProvidersRP represents MagmaSC relative path.
	// Used to list all registered in the blockchain providers.
	GetAllProvidersRP = "/allProviders"
)

const (
	// AcknowledgmentPrefix represents prefix to save into storage.
	AcknowledgmentPrefix = "ackn"

	// AcknowledgmentActivePrefixPart represents prefix part to save into storage.
	AcknowledgmentActivePrefixPart = "act"

	// AcknowledgmentRP represents MagmaSC relative path.
	// Used to retrieve accepted acknowledgment.
	AcknowledgmentRP = "/acknowledgmentAccepted"

	// IsAcknowledgmentExistRP represents MagmaSC relative path.
	// Used to check existing of acknowledgment.
	IsAcknowledgmentExistRP = "/acknowledgmentExist"

	// VerifyAcknowledgmentAcceptedRP represents MagmaSC relative path.
	// Used to verify accepting Provider's terms by Consumer.
	VerifyAcknowledgmentAcceptedRP = "/acknowledgmentAcceptedVerify"
)

const (
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

	// consumerType contents a value of consumer node type.
	consumerType = "consumer"
)

const (
	// TermsExpiredDuration represents value for
	// minimal duration of provider terms that will pass check it's expired.
	TermsExpiredDuration = 1 * 60 // 1 minute

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

	// ProviderSessionInitFuncName represents MagmaSC function.
	// Used for initializing session by a provider.
	ProviderSessionInitFuncName = "provider_session_init"

	// ProviderFetchRP represents MagmaSC relative path.
	// Used to fetch provider info.
	ProviderFetchRP = "/providerFetch"

	// ProviderRegisteredRP represents MagmaSC relative path.
	// Used to fetch provider registered info.
	ProviderRegisteredRP = "/providerExist"

	// providerType contents a value of provider node type.
	providerType = "provider"
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
