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
		wantErr   bool
	}{
		{
			name:      "valid plaintext",
			key:       "passphrase1111111111111111111111",
			plaintext: "glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp",
			wantErr:   false,
		},
		{
			name:      "empty key",
			key:       "",
			plaintext: "glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp",
			wantErr:   true,
		},
		{
			name:      "empty plaintext",
			key:       "passphrase1111111111111111111111",
			plaintext: "",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key := []byte(tc.key)
			plaintext := []byte(tc.plaintext)

			// Encrypt plaintext
			ciphertext, err := ScryptEncrypt(key, plaintext)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, ciphertext)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ciphertext)

				// Decrypt ciphertext
				decryptedText, err := ScryptDecrypt(key, ciphertext)
				require.NoError(t, err)
				require.Equal(t, plaintext, decryptedText)
			}
		})
	}
}
