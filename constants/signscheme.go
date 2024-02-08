package constants

type SignScheme string

const (
	ED25519 SignScheme = "ed25519"
	BLS0CHAIN SignScheme = "bls0chain"
)

func (s SignScheme) String() string {
	return string(s)
}

