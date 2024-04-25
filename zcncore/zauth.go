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
}

// CallZauthSetup calls the zauth setup endpoint
func CallZauthSetup(serverAddr string, token string, splitWallet SplitWallet) error {
	// Add your code here
	endpoint := serverAddr + "/setup"
	wData, err := json.Marshal(splitWallet)
	if err != nil {
		return errors.Wrap(err, "failed to marshal split wallet")
	}

	fmt.Println("call zauth setup:", endpoint)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(wData))
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}

	fmt.Println("split wallet:", splitWallet)

	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("X-Client-ID", splitWallet.ClientID)
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

func CallZauthSetupString(serverAddr, splitWallet string) error {
	// Add your code here
	endpoint := serverAddr + "/setup"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(splitWallet)))
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
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

// ZauthSignTxn returns a function that sends a txn signing request to the zauth server
func ZauthSignTxn(serverAddr string) sys.AuthorizeFunc {
	return func(msg string) (string, error) {
		req, err := http.NewRequest("POST", serverAddr+"/sign/txn", bytes.NewBuffer([]byte(msg)))
		if err != nil {
			return "", errors.Wrap(err, "failed to create HTTP request")
		}
		req.Header.Set("Content-Type", "application/json")

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
		req.Header.Set("Content-Type", "application/json")
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
