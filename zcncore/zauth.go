package zcncore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/pkg/errors"
)

// AvailableRestrictions represents supported restrictions mapping.
var AvailableRestrictions = map[string][]string{
	"token_transfers": {"transfer"},
	"allocation_file_operations": {
		"read_redeem",
		"commit_connection",
	},
	"allocation_storage_operations": {
		"new_allocation_request",
		"update_allocation_request",
		"finalize_allocation",
		"cancel_allocation",
		"add_free_storage_assigner",
		"free_allocation_request",
	},
	"allocation_token_operations": {
		"read_pool_lock",
		"read_pool_unlock",
		"write_pool_lock",
	},
	"storage_rewards": {
		"collect_reward",
		"stake_pool_lock",
		"stake_pool_unlock",
	},
	"storage_operations": {
		"challenge_response",
		"add_validator",
		"add_blobber",
		"blobber_health_check",
		"validator_health_check",
	},
	"storage_management": {
		"kill_blobber",
		"kill_validator",
		"shutdown_blobber",
		"shutdown_validator",
		"update_blobber_settings",
		"update_validator_settings",
	},
	"miner_operations": {
		"add_miner",
		"add_sharder",
		"miner_health_check",
		"sharder_health_check",
		"contributeMpk",
		"shareSignsOrShares",
		"wait",
		"sharder_keep",
	},
	"miner_management_operations": {
		"delete_miner",
		"delete_sharder",
		"update_miner_settings",
		"kill_miner",
		"kill_sharder",
	},
	"miner_financial_operations": {
		"addToDelegatePool",
		"deleteFromDelegatePool",
		"collect_reward",
	},
	"token_bridging": {
		"mint",
		"burn",
	},
	"authorizer_management_operations": {
		"delete-authorizer",
	},
	"authorizer_operations": {
		"add-authorizer",
		"authorizer-health-check",
		"add-to-delegate-pool",
		"delete-from-delegate-pool",
	},
}

type createKeyRequest struct {
	Restrictions []string `json:"restrictions"`
}

type updateRestrictionsRequest struct {
	Restrictions []string `json:"restrictions"`
}

type AuthMessage struct {
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
	ClientID  string `json:"client_id"`
}

type AuthResponse struct {
	Sig string `json:"sig"`
}

func CallZauthRetreiveKey(serverAddr, token, clientID, peerPublicKey string) (string, error) {
	endpoint := fmt.Sprintf("%s/key/%s", serverAddr, clientID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Peer-Public-Key", peerPublicKey)
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

func CallZauthRevoke(serverAddr, token, clientID, peerPublicKey string) error {
	endpoint := serverAddr + "/revoke/" + clientID
	endpoint += "?peer_public_key=" + peerPublicKey
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		if len(errMsg) > 0 {
			return errors.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
		}

		return errors.Errorf("code: %d", resp.StatusCode)
	}

	return nil
}

func CallZauthDelete(serverAddr, token, clientID string) error {
	endpoint := serverAddr + "/delete/" + clientID
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		if len(errMsg) > 0 {
			return errors.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
		}

		return errors.Errorf("code: %d", resp.StatusCode)
	}

	return nil
}

func CallZvaultNewWallet(serverAddr, token string) (string, error) {
	endpoint := serverAddr + "/wallet"

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		if len(errMsg) > 0 {
			return "", errors.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
		}

		return "", errors.Errorf("code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

func CallZvaultNewSplit(serverAddr, token, clientID string, restrictions []string) (string, error) {
	endpoint := serverAddr + "/key/" + clientID

	data, err := json.Marshal(createKeyRequest{
		Restrictions: restrictions,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to serialize request")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		if len(errMsg) > 0 {
			return "", errors.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
		}

		return "", errors.Errorf("code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

func CallZvaultUpdateRestrictions(serverAddr, token, clientID, peerPublicKey string, restrictions []string) (string, error) {
	endpoint := serverAddr + "/restrictions/" + clientID

	data, err := json.Marshal(updateRestrictionsRequest{
		Restrictions: restrictions,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to serialize request")
	}

	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(data))
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Peer-Public-Key", peerPublicKey)
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		if len(errMsg) > 0 {
			return "", errors.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
		}

		return "", errors.Errorf("code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

func CallZvaultStoreKeyString(serverAddr, token, privateKey string) (string, error) {
	// Add your code here
	endpoint := serverAddr + "/store"

	reqData := struct {
		PrivateKey string `json:"private_key"`
	}{
		PrivateKey: privateKey,
	}

	var buff bytes.Buffer

	encoder := json.NewEncoder(&buff)

	err := encoder.Encode(reqData)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	var req *http.Request

	req, err = http.NewRequest("POST", endpoint, &buff)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		if len(errMsg) > 0 {
			return "", errors.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
		}

		return "", errors.Errorf("code: %d", resp.StatusCode)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

func CallZvaultRetrieveKeys(serverAddr, token, clientID string) (string, error) {
	// Add your code here
	endpoint := fmt.Sprintf("%s/keys/%s", serverAddr, clientID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

func CallZvaultDeletePrimaryKey(serverAddr, token, clientID string) error {
	// Add your code here
	endpoint := serverAddr + "/delete/" + clientID
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	return nil
}

func CallZvaultRevokeKey(serverAddr, token, clientID, publicKey string) error {
	// Add your code here
	endpoint := fmt.Sprintf("%s/revoke/%s?public_key=%s", serverAddr, clientID, publicKey)
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	return nil
}

func CallZvaultRetrieveWallets(serverAddr, token string) (string, error) {
	// Add your code here
	endpoint := fmt.Sprintf("%s/wallets", serverAddr)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

func CallZvaultRetrieveSharedWallets(serverAddr, token string) (string, error) {
	// Add your code here
	endpoint := fmt.Sprintf("%s/wallets/shared", serverAddr)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("code: %d, err: %s", resp.StatusCode, string(errMsg))
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	return string(d), nil
}

// ZauthSignTxn returns a function that sends a txn signing request to the zauth server
func ZauthSignTxn(serverAddr string) sys.AuthorizeFunc {
	return func(msg string) (string, error) {
		req, err := http.NewRequest("POST", serverAddr+"/sign/txn", bytes.NewBuffer([]byte(msg)))
		if err != nil {
			return "", errors.Wrap(err, "failed to create HTTP request")
		}
		req.Header.Set("Content-Type", "application/json")
		c := client.GetClient()
		pubkey := c.Keys[0].PublicKey
		req.Header.Set("X-Peer-Public-Key", pubkey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", errors.Wrap(err, "failed to send HTTP request")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			rsp, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", errors.Wrap(err, "failed to read response body")
			}

			return "", errors.Errorf("unexpected status code: %d, res: %s", resp.StatusCode, string(rsp))
		}

		d, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "failed to read response body")
		}

		return string(d), nil
	}
}

func ZauthAuthCommon(serverAddr string) sys.AuthorizeFunc {
	return func(msg string) (string, error) {
		req, err := http.NewRequest("POST", serverAddr+"/sign/msg", bytes.NewBuffer([]byte(msg)))
		if err != nil {
			return "", errors.Wrap(err, "failed to create HTTP request")
		}

		c := client.GetClient()
		pubkey := c.Keys[0].PublicKey
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Peer-Public-Key", pubkey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", errors.Wrap(err, "failed to send HTTP request")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			rsp, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", errors.Wrap(err, "failed to read response body")
			}

			return "", errors.Errorf("unexpected status code: %d, res: %s", resp.StatusCode, string(rsp))
		}

		d, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "failed to read response body")
		}

		return string(d), nil
	}
}
