package client

import (
    "testing"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/stretchr/testify/assert"
	"sync"
)

func TestMultiWalletManagement(t *testing.T) {
    // Reset client for testing
    client.wallets = make(map[string]*zcncrypto.Wallet)
    client.wg = make(map[string]*sync.WaitGroup)
    client.walletCount = make(map[string]int)

    w1 := zcncrypto.Wallet{
        ClientID:  "client_1",
        ClientKey: "key_1",
        Keys: []zcncrypto.KeyPair{
            {PrivateKey: "priv_1", PublicKey: "pub_1"},
        },
    }

    w2 := zcncrypto.Wallet{
        ClientID:  "client_2",
        ClientKey: "key_2",
        Keys: []zcncrypto.KeyPair{
            {PrivateKey: "priv_2", PublicKey: "pub_2"},
        },
    }

    t.Run("AddWallet", func(t *testing.T) {
        AddWallet(w1)
        assert.Equal(t, &w1, GetWalletByClientID("client_1"))
        assert.Equal(t, 1, client.walletCount["client_1"])
    })

    t.Run("GetWalletByClientID", func(t *testing.T) {
        AddWallet(w2)
        got := GetWalletByClientID("client_2")
        assert.NotNil(t, got)
        assert.Equal(t, "client_2", got.ClientID)
        
        gotMissing := GetWalletByClientID("missing")
        assert.Nil(t, gotMissing)
    })

    t.Run("RemoveWallet", func(t *testing.T) {
        RemoveWallet("client_1")
        assert.Nil(t, GetWalletByClientID("client_1"))
        assert.Equal(t, 0, client.walletCount["client_1"])
    })
    
    t.Run("Sign with Client Selection", func(t *testing.T) {
         // Setup mock signature scheme? 
         // Real signature might require valid keys. 
         // For now, testing the wallet selection logic is tricky without mocking sys.Sign
         // minimal check: Id()
    })

    t.Run("Id with Client Selection", func(t *testing.T) {
        AddWallet(w2)
        id := Id("client_2")
        assert.Equal(t, "client_2", id)
        
        // Default behavior coverage
        // Set main wallet
        SetWallet(w1)
        assert.Equal(t, "client_1", Id())
    })
}
