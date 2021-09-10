package magmasc

import (
	"encoding/json"

	"github.com/0chain/gosdk/zmagmacore/http"
)

// GetAllConsumers makes smart contract rest api call to magma smart contract
// GetAllConsumersRP rest point to retrieve all registered consumer.Consumer.
func GetAllConsumers() ([]Consumer, error) {
	resp, err := http.MakeSCRestAPICall(Address, GetAllConsumersRP, nil)
	if err != nil {
		return nil, err
	}

	consumers := make([]Consumer, 0)
	if err = json.Unmarshal(resp, &consumers); err != nil {
		return nil, err
	}

	return consumers, err
}

// GetAllProviders makes smart contract rest api call to magma smart contract
// GetAllProvidersRP rest point to retrieve all registered provider.Provider.
func GetAllProviders() ([]Provider, error) {
	resp, err := http.MakeSCRestAPICall(Address, GetAllProvidersRP, nil)
	if err != nil {
		return nil, err
	}

	providers := make([]Provider, 0)
	if err = json.Unmarshal(resp, &providers); err != nil {
		return nil, err
	}

	return providers, err
}

// RequestAcknowledgment makes smart contract rest api call to magma smart contract
// AcknowledgmentRP rest point to retrieve Acknowledgment.
func RequestAcknowledgment(sessionID string) (*Acknowledgment, error) {
	params := map[string]string{
		"id": sessionID,
	}

	blob, err := http.MakeSCRestAPICall(Address, AcknowledgmentRP, params)
	if err != nil {
		return nil, err
	}

	ackn := Acknowledgment{}
	if err = json.Unmarshal(blob, &ackn); err != nil {
		return nil, err
	}

	return &ackn, nil
}

// IsAcknowledgmentExist makes smart contract rest api call to magma smart contract
// IsAcknowledgmentExistRP rest point to ensure that Acknowledgment with provided session ID exist in the blockchain.
func IsAcknowledgmentExist(sessionID string) (bool, error) {
	params := map[string]string{
		"id": sessionID,
	}

	blob, err := http.MakeSCRestAPICall(Address, IsAcknowledgmentExistRP, params)
	if err != nil {
		return false, err
	}

	var exist bool
	if err = json.Unmarshal(blob, &exist); err != nil {
		return false, err
	}

	return exist, nil
}

// VerifyAcknowledgmentAccepted makes smart contract rest api call to magma smart contract
// VerifyAcknowledgmentAcceptedRP rest point to ensure that Acknowledgment with provided IDs was accepted.
func VerifyAcknowledgmentAccepted(sessionID, accessPointID, consumerExtID, providerExtID string) (*Acknowledgment, error) {
	params := map[string]string{
		"session_id":      sessionID,
		"access_point_id": accessPointID,
		"provider_ext_id": providerExtID,
		"consumer_ext_id": consumerExtID,
	}

	blob, err := http.MakeSCRestAPICall(Address, VerifyAcknowledgmentAcceptedRP, params)
	if err != nil {
		return nil, err
	}

	ackn := Acknowledgment{}
	if err = json.Unmarshal(blob, &ackn); err != nil {
		return nil, err
	}

	return &ackn, nil
}

// ConsumerFetch makes smart contract rest api call to magma smart contract
// ConsumerFetchRP rest point to fetch Consumer info.
func ConsumerFetch(id string) (*Consumer, error) {
	params := map[string]string{
		"ext_id": id,
	}

	blob, err := http.MakeSCRestAPICall(Address, ConsumerFetchRP, params)
	if err != nil {
		return nil, err
	}

	cons := Consumer{}
	if err = json.Unmarshal(blob, &cons); err != nil {
		return nil, err
	}

	return &cons, nil
}

// ProviderFetch makes smart contract rest api call to magma smart contract
// ProviderFetchRP rest point to fetch Provider info.
func ProviderFetch(id string) (*Provider, error) {
	params := map[string]string{
		"ext_id": id,
	}

	blob, err := http.MakeSCRestAPICall(Address, ProviderFetchRP, params)
	if err != nil {
		return nil, err
	}

	prov := Provider{}
	if err = json.Unmarshal(blob, &prov); err != nil {
		return nil, err
	}

	return &prov, nil
}

// IsConsumerRegisteredRP makes smart contract rest api call to magma smart contract
// ConsumerRegisteredRP rest point to check registration of the consumer with provided external ID.
func IsConsumerRegisteredRP(extID string) (bool, error) {
	params := map[string]string{
		"ext_id": extID,
	}
	registeredByt, err := http.MakeSCRestAPICall(Address, ConsumerRegisteredRP, params)
	if err != nil {
		return false, err
	}

	var registered bool
	if err := json.Unmarshal(registeredByt, &registered); err != nil {
		return false, err
	}

	return registered, nil
}

// IsProviderRegisteredRP makes smart contract rest api call to magma smart contract
// ProviderRegisteredRP rest point to check registration of the provider with provided external ID.
func IsProviderRegisteredRP(extID string) (bool, error) {
	params := map[string]string{
		"ext_id": extID,
	}
	registeredByt, err := http.MakeSCRestAPICall(Address, ProviderRegisteredRP, params)
	if err != nil {
		return false, err
	}

	var registered bool
	if err := json.Unmarshal(registeredByt, &registered); err != nil {
		return false, err
	}

	return registered, nil
}
