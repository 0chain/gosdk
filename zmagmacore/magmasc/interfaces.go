package magmasc

func (m *Consumer) ExternalID() string {
	return m.ExtID
}

func (m *Consumer) FetchNodeRP() string {
	return ConsumerFetchRP
}

func (m *Consumer) IsNodeRegisteredRP() string {
	return ConsumerRegisteredRP
}

func (m *Consumer) RegistrationFuncName() string {
	return ConsumerRegisterFuncName
}

func (m *Consumer) UpdateNodeFuncName() string {
	return ConsumerUpdateFuncName
}

func (m *Provider) ExternalID() string {
	return m.ExtID
}

func (m *Provider) FetchNodeRP() string {
	return ProviderFetchRP
}

func (m *Provider) IsNodeRegisteredRP() string {
	return ProviderRegisteredRP
}

func (m *Provider) RegistrationFuncName() string {
	return ProviderRegisterFuncName
}

func (m *Provider) UpdateNodeFuncName() string {
	return ProviderUpdateFuncName
}
