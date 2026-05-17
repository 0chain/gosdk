package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZauthAuthCommon_RecoveryOnResourceNotFound(t *testing.T) {
	var signCallCount int32

	// Mock zauth server: first call returns "resource not found", second succeeds
	zauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&signCallCount, 1)
		if count == 1 {
			// First call: keys not found
			http.Error(w, "keys of client id: test123: resource not found", http.StatusInternalServerError)
			return
		}
		// Second call (after recovery): success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`"signed_message_data"`))
	}))
	defer zauthServer.Close()

	// Track recovery calls
	var recoveryCalled int32
	ZauthKeyRecoveryFunc = func(clientID string) error {
		atomic.AddInt32(&recoveryCalled, 1)
		assert.Equal(t, "test_client_id", clientID)
		return nil
	}
	defer func() { ZauthKeyRecoveryFunc = nil }()

	// Set up a test client so GetClient() works
	SetWallet(zcncrypto.Wallet{ClientID: "test_client_id", ClientKey: "test_client_key", Keys: []zcncrypto.KeyPair{{PublicKey: "test_pub_key", PrivateKey: "test_priv_key"}}})

	// Build the auth message
	authMsg := struct {
		Hash      string `json:"hash"`
		Signature string `json:"signature"`
		ClientID  string `json:"client_id"`
	}{
		Hash:      "test_hash",
		Signature: "test_sig",
		ClientID:  "test_client_id",
	}
	msgBytes, _ := json.Marshal(authMsg)

	// Call ZauthAuthCommon
	authFn := ZauthAuthCommon(zauthServer.URL)
	result, err := authFn(string(msgBytes))

	// Verify recovery was triggered and retry succeeded
	require.NoError(t, err)
	assert.Equal(t, "\"signed_message_data\"", result)
	assert.Equal(t, int32(1), atomic.LoadInt32(&recoveryCalled), "recovery should be called once")
	assert.Equal(t, int32(2), atomic.LoadInt32(&signCallCount), "sign/msg should be called twice (original + retry)")
}

func TestZauthAuthCommon_NoRecoveryOnOtherErrors(t *testing.T) {
	// Mock zauth server: returns a different error (not "resource not found")
	zauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid signature", http.StatusBadRequest)
	}))
	defer zauthServer.Close()

	var recoveryCalled int32
	ZauthKeyRecoveryFunc = func(clientID string) error {
		atomic.AddInt32(&recoveryCalled, 1)
		return nil
	}
	defer func() { ZauthKeyRecoveryFunc = nil }()

	SetWallet(zcncrypto.Wallet{ClientID: "test_client_id", ClientKey: "test_client_key", Keys: []zcncrypto.KeyPair{{PublicKey: "test_pub_key", PrivateKey: "test_priv_key"}}})

	authMsg := struct {
		Hash      string `json:"hash"`
		Signature string `json:"signature"`
		ClientID  string `json:"client_id"`
	}{
		Hash:      "test_hash",
		Signature: "test_sig",
		ClientID:  "test_client_id",
	}
	msgBytes, _ := json.Marshal(authMsg)

	authFn := ZauthAuthCommon(zauthServer.URL)
	_, err := authFn(string(msgBytes))

	// Should error but NOT trigger recovery
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
	assert.Equal(t, int32(0), atomic.LoadInt32(&recoveryCalled), "recovery should NOT be called for non-resource-not-found errors")
}

func TestZauthAuthCommon_NoRecoveryWhenFuncNil(t *testing.T) {
	// Mock zauth server: returns "resource not found"
	zauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "resource not found", http.StatusInternalServerError)
	}))
	defer zauthServer.Close()

	ZauthKeyRecoveryFunc = nil

	SetWallet(zcncrypto.Wallet{ClientID: "test_client_id", ClientKey: "test_client_key", Keys: []zcncrypto.KeyPair{{PublicKey: "test_pub_key", PrivateKey: "test_priv_key"}}})

	authMsg := struct {
		Hash      string `json:"hash"`
		Signature string `json:"signature"`
		ClientID  string `json:"client_id"`
	}{
		Hash:      "test_hash",
		Signature: "test_sig",
		ClientID:  "test_client_id",
	}
	msgBytes, _ := json.Marshal(authMsg)

	authFn := ZauthAuthCommon(zauthServer.URL)
	_, err := authFn(string(msgBytes))

	// Should error without panic (no recovery func set)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource not found")
}
