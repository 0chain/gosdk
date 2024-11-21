package constants

// ContextKey type for key used to store values into context
type ContextKey string

const (
	// ContextKeyAllocation represents key for context value for allocation
	ContextKeyAllocation ContextKey = "allocation"

	// ContextKeyAllocationObject represents key for context value for allocation object
	ContextKeyAllocationObject ContextKey = "allocation_object"

	// ContextKeyClient represents key for context value for client
	ContextKeyClient ContextKey = "client"
	
	// ContextKeyClientKey represents key for context value for client key
	ContextKeyClientKey ContextKey = "client_key"

	// ContextKeyClientID represents key for context value for allocation id
	ContextKeyAllocationID ContextKey = "allocation_id"

	// ContextKeyClientSignatureHeaderKey represents key for context value passed with common.ClientSignatureHeader request header.
	ContextKeyClientSignatureHeaderKey   ContextKey = "signature"

	// ContextKeyClientSignatureHeaderV2Key represents key for context value passed with common.ClientSignatureHeaderV2 request header.
	ContextKeyClientSignatureHeaderV2Key ContextKey = "signature_v2"
)
