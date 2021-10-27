package constants

/*ContextKey - type for key used to store values into context */
type ContextKey string

const (
	// ContextKeyAllocation
	ContextKeyAllocation ContextKey = "allocation"
	// ContextKeyAllocationObject
	ContextKeyAllocationObject ContextKey = "allocation_object"
	// ContextKeyClient
	ContextKeyClient ContextKey = "client"
	// ContextKeyClientKey
	ContextKeyClientKey ContextKey = "client_key"

	// ContextKeyClientSignatureHeaderKey represents key for context value passed with common.ClientSignatureHeader request header.
	ContextKeyClientSignatureHeaderKey ContextKey = "signature"
)
