package zbox

import (
	"fmt"
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"regexp"
	"strconv"
)

func GetNumber(value string) int {
	re := regexp.MustCompile("[0-9]+")
	submatchall := re.FindAllString(value, -1)
	for _, element := range submatchall {
		res, _ := strconv.Atoi(element)
		return res
	}
	return -1
}

func Sign(clientId string) (string, error) {
	if len(clientId) == 0 {
		return "", fmt.Errorf("null client ID for signing")
	}
	return client.Sign(clientId)
}

func GetClientEncryptedPublicKey() (string, error) {
	return sdk.GetClientEncryptedPublicKey()
}
