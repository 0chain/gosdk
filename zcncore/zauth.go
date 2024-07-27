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

// SplitWallet represents wallet info for split wallet
// The client id and client key are the same as the primary wallet client id and client key
type SplitWallet struct {
	ClientID      string `json:"client_id"`
	ClientKey     string `json:"client_key"`
	PublicKey     string `json:"public_key"`
	PrivateKey    string `json:"private_key"`
	PeerPublicKey string `json:"peer_public_key"`
	IsRevoked     bool   `json:"is_revoked"`
	ExpiredAt     int64  `json:"expired_at"`
}

// CallZauthSetup calls the zauth setup endpoint
func CallZauthSetup(serverAddr string, token string, splitWallet SplitWallet) error {
	// Add your code here
	endpoint := serverAddr + "/setup"
	wData, err := json.Marshal(splitWallet)
	if err != nil {
		return errors.Wrap(err, "failed to marshal split wallet")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(wData))
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

	var rsp struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rsp); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	if rsp.Result != "success" {
		return errors.New("failed to setup zauth server")
	}

	return nil
}

func CallZauthRevoke(serverAddr, token, clientID, publicKey string) error {
	endpoint := serverAddr + "/revoke/" + clientID
	endpoint += "?peer_public_key=" + publicKey
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

	var rsp struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rsp); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	if rsp.Result != "success" {
		return errors.New("failed to setup zauth server")
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

	var rsp struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rsp); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}

	if rsp.Result != "success" {
		return errors.New("failed to setup zauth server")
	}

	return nil
}

func CallZvaultNewWalletString(serverAddr, token, clientID string) (string, error) {
	// Add your code here
	endpoint := serverAddr + "/generate"
	if clientID != "" {
		endpoint = endpoint + "/" + clientID
	}

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	fmt.Println("new wallet endpoint:", endpoint)
	fmt.Println("new wallet: serverAddr:", serverAddr)
	fmt.Println("new wallet: clientID:", clientID)

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

	fmt.Println("call zvault /store:", endpoint)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jwt-Token", token)

	fmt.Println(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())

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

	fmt.Println("call zvault /keys:", endpoint)
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

	fmt.Println("call zvault /delete:", endpoint)
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

	fmt.Println("call zvault /revoke:", endpoint)
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

	fmt.Println("call zvault /keys:", endpoint)
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

	fmt.Println("call zvault /keys:", endpoint)
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
		fmt.Println("zvault sign txn - in sign txn...")
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
		// return func(msg string) (string, error) {
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

type AuthMessage struct {
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
	ClientID  string `json:"client_id"`
}

type AuthResponse struct {
	Sig string `json:"sig"`
}

func ZauthSignMsg(serverAddr string) sys.SignFunc {
	return func(hash string, signatureScheme string, keys []sys.KeyPair) (string, error) {
		sig, err := SignWithKey(keys[0].PrivateKey, hash)
		if err != nil {
			return "", err
		}

		data, err := json.Marshal(AuthMessage{
			Hash:      hash,
			Signature: sig,
			ClientID:  client.GetClient().ClientID,
		})
		if err != nil {
			return "", err
		}

		// fmt.Println("auth - sys.AuthCommon:", sys.AuthCommon)
		if sys.AuthCommon == nil {
			return "", errors.New("authCommon is not set")
		}

		rsp, err := sys.AuthCommon(string(data))
		if err != nil {
			return "", err
		}

		var ar AuthResponse
		err = json.Unmarshal([]byte(rsp), &ar)
		if err != nil {
			return "", err
		}

		return AddSignature(client.GetClientPrivateKey(), ar.Sig, hash)
	}
}
