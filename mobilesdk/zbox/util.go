package zbox

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/0chain/gosdk/zcncore"
)

// GetClientEncryptedPublicKey - getting client encrypted pub key
func GetClientEncryptedPublicKey() (string, error) {
	return sdk.GetClientEncryptedPublicKey()
}

func TokensToEth(tokens int64) string {
	return fmt.Sprintf("%f", zcncore.TokensToEth(tokens))
}

func GEthToTokens(tokens int64) string {
	return fmt.Sprintf("%f", zcncore.GTokensToEth(tokens))
}

// ConvertZcnTokenToETH - converting Zcn tokens to Eth
func ConvertZcnTokenToETH(token float64) (string, error) {
	res, err := zcncore.ConvertZcnTokenToETH(token)
	return fmt.Sprintf("%f", res), err
}

// SuggestEthGasPrice - return back suggested price for gas
func SuggestEthGasPrice() (string, error) {
	res, err := zcncore.SuggestEthGasPrice()
	return strconv.FormatInt(res, 10), err
}

// Encrypt - encrypting text with key
func Encrypt(key, text string) (string, error) {
	keyBytes := []byte(key)
	textBytes := []byte(text)
	response, err := zboxutil.Encrypt(keyBytes, textBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(response), nil
}

// Decrypt - decrypting text with key
func Decrypt(key, text string) (string, error) {
	keyBytes := []byte(key)
	textBytes, _ := hex.DecodeString(text)
	response, err := zboxutil.Decrypt(keyBytes, textBytes)
	if err != nil {
		return "", err
	}
	return string(response), nil
}

// GetNetwork - get current network
func GetNetwork() (string, error) {
	networkDetails := sdk.GetNetwork()
	networkDetailsBytes, err := json.Marshal(networkDetails)
	if err != nil {
		return "", err
	}
	return string(networkDetailsBytes), nil
}

// GetBlobbers - get list of blobbers
func GetBlobbers() (string, error) {
	blobbers, err := sdk.GetBlobbers()
	if err != nil {
		return "", err
	}

	blobbersBytes, err := json.Marshal(blobbers)
	if err != nil {
		return "", err
	}
	return string(blobbersBytes), nil
}

// Sign - sign hash
func Sign(hash string) (string, error) {
	if len(hash) == 0 {
		return "", fmt.Errorf("null sign")
	}
	return client.Sign(hash)
}

// VerifySignature - verify message with signature
func VerifySignature(signature string, msg string) (bool, error) {
	return client.VerifySignature(signature, msg)
}

func GetNumber(value string) int {
	re := regexp.MustCompile("[0-9]+")
	submatchall := re.FindAllString(value, -1)
	for _, element := range submatchall {
		res, _ := strconv.Atoi(element)
		return res
	}
	return -1
}
