package magmasc

// ExternalID represents simple getter for ExtID.
func (m *Consumer) ExternalID() string {
	return m.ExtID
}

// FetchNodeRP returns name of magma sc rest point used for fetching consumer's node info.
func (m *Consumer) FetchNodeRP() string {
	return ConsumerFetchRP
}

// IsNodeRegisteredRP returns name of magma sc rest point used for checking consumer's node registration.
func (m *Consumer) IsNodeRegisteredRP() string {
	return ConsumerRegisteredRP
}

// RegistrationFuncName returns name of magma sc function used for consumer's node registration.
func (m *Consumer) RegistrationFuncName() string {
	return ConsumerRegisterFuncName
}

// UpdateNodeFuncName returns name of magma sc function used for consumer's node updating.
func (m *Consumer) UpdateNodeFuncName() string {
	return ConsumerUpdateFuncName
}

// ExternalID represents simple getter for ExtID.
func (m *Provider) ExternalID() string {
	return m.ExtID
}

// FetchNodeRP returns name of magma sc rest point used for fetching provider's node info.
func (m *Provider) FetchNodeRP() string {
	return ProviderFetchRP
}

// IsNodeRegisteredRP returns name of magma sc rest point used for checking provider's node registration.
func (m *Provider) IsNodeRegisteredRP() string {
	return ProviderRegisteredRP
}

// RegistrationFuncName returns name of magma sc function used for provider's node registration.
func (m *Provider) RegistrationFuncName() string {
	return ProviderRegisterFuncName
}

// UpdateNodeFuncName returns name of magma sc function used for provider's node updating.
func (m *Provider) UpdateNodeFuncName() string {
	return ProviderUpdateFuncName
}
