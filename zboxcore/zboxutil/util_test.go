package zboxutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScryptEncryption(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		key       string
		plaintext string
	}{
		{
			name:      "valid plaintext",
			key:       "passphrase1111111111111111111111",
			plaintext: "glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp",
		},
		{
			name:      "empty key",
			key:       "",
			plaintext: "glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp",
		},
		{
			name:      "empty plaintext",
			key:       "passphrase1111111111111111111111",
			plaintext: "",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key := []byte(tc.key)
			plaintext := []byte(tc.plaintext)

			// Encrypt plaintext
			ciphertext, err := ScryptEncrypt(key, plaintext)
			require.NoError(t, err)
			require.NotEmpty(t, ciphertext)

			// Decrypt ciphertext
			decryptedText, err := ScryptDecrypt(key, ciphertext)
			require.NoError(t, err)
			require.Equal(t, plaintext, decryptedText)
		})
	}
}
