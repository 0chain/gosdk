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

// RequestSession makes smart contract rest api call to magma smart contract
// SessionRP rest point to retrieve Session.
func RequestSession(sessionID string) (*Session, error) {
	params := map[string]string{
		"id": sessionID,
	}

	blob, err := http.MakeSCRestAPICall(Address, SessionRP, params)
	if err != nil {
		return nil, err
	}

	session := Session{}
	if err = json.Unmarshal(blob, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// IsSessionExist makes smart contract rest api call to magma smart contract
// IsSessionExistRP rest point to ensure that Session with provided session ID exist in the blockchain.
func IsSessionExist(sessionID string) (bool, error) {
	params := map[string]string{
		"id": sessionID,
	}

	blob, err := http.MakeSCRestAPICall(Address, IsSessionExistRP, params)
	if err != nil {
		return false, err
	}

	var exist bool
	if err = json.Unmarshal(blob, &exist); err != nil {
		return false, err
	}

	return exist, nil
}

// VerifySessionAccepted makes smart contract rest api call to magma smart contract
// VerifySessionAcceptedRP rest point to ensure that Session with provided IDs was accepted.
func VerifySessionAccepted(sessionID, accessPointID, consumerExtID, providerExtID string) (*Session, error) {
	params := map[string]string{
		"session_id":      sessionID,
		"access_point_id": accessPointID,
		"provider_ext_id": providerExtID,
		"consumer_ext_id": consumerExtID,
	}

	blob, err := http.MakeSCRestAPICall(Address, VerifySessionAcceptedRP, params)
	if err != nil {
		return nil, err
	}

	session := Session{}
	if err = json.Unmarshal(blob, &session); err != nil {
		return nil, err
	}

	return &session, nil
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

// AccessPointFetch makes smart contract rest api call to magma smart contract
// AccessPointFetchRP rest point to fetch AccessPoint info.
func AccessPointFetch(id string) (*AccessPoint, error) {
	params := map[string]string{
		"id": id,
	}

	blob, err := http.MakeSCRestAPICall(Address, AccessPointFetchRP, params)
	if err != nil {
		return nil, err
	}

	aP := AccessPoint{}
	if err = json.Unmarshal(blob, &aP); err != nil {
		return nil, err
	}

	return &aP, nil
}

// ProviderMinStakeFetch makes smart contract rest api call to magma smart contract
// ProviderFetchRP rest point to fetch Provider info.
func ProviderMinStakeFetch() (int64, error) {
	blob, err := http.MakeSCRestAPICall(Address, ProviderMinStakeFetchRP, nil)
	if err != nil {
		return 0, err
	}

	var minStake int64
	if err = json.Unmarshal(blob, &minStake); err != nil {
		return 0, err
	}

	return minStake, nil
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

// IsAccessPointRegisteredRP makes smart contract rest api call to magma smart contract
// AccessPointRegisteredRP rest point to check registration of the access point with provided ID.
func IsAccessPointRegisteredRP(ID string) (bool, error) {
	params := map[string]string{
		"id": ID,
	}
	registeredByt, err := http.MakeSCRestAPICall(Address, AccessPointRegisteredRP, params)
	if err != nil {
		return false, err
	}

	var registered bool
	if err := json.Unmarshal(registeredByt, &registered); err != nil {
		return false, err
	}

	return registered, nil
}

// AccessPointMinStakeFetch makes smart contract rest api call to magma smart contract
// AccessPointMinStakeFetchRP rest point to fetch configured min stake value.
func AccessPointMinStakeFetch() (int64, error) {
	minStakeByte, err := http.MakeSCRestAPICall(Address, AccessPointMinStakeFetchRP, nil)
	if err != nil {
		return 0, err
	}

	var minStake int64
	if err := json.Unmarshal(minStakeByte, &minStake); err != nil {
		return 0, err
	}

	return minStake, nil
}

// FetchBillingRatio makes smart contract rest api call to magma smart contract
// FetchBillingRatioRP rest point to fetch configured billing ratio.
func FetchBillingRatio() (int64, error) {
	billingRatioByt, err := http.MakeSCRestAPICall(Address, FetchBillingRatioRP, nil)
	if err != nil {
		return 0, err
	}

	var billingRatio int64
	if err := json.Unmarshal(billingRatioByt, &billingRatio); err != nil {
		return 0, err
	}

	return billingRatio, nil
}
