package sdk

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	privateKey      = "41729ed8d82f782646d2d30b9719acfd236842b9b6e47fee12b7bdbd05b35122"
	publicKey       = "1f495df9605a4479a7dd6e5c7a78caf9f9d54e3a40f62a3dd68ed377115fe614d8acf0c238025f67a85163b9fbf31d10fbbb4a551d1cf00119897edf18b1841c"
	signatureScheme = "bls0chain"
)

func TestSign(t *testing.T) {

	fields := []string{"input1", "input2", "input3"}

	signature, err := SignRequest(privateKey, signatureScheme, strings.Join(fields, ":"))

	require.NoError(t, err)

	ok, err := VerifySignature(publicKey, signatureScheme, strings.Join(fields, ":"), signature)
	require.NoError(t, err)
	require.True(t, ok)

}
