package authorizer

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/core/client"

	"github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/core/zcncrypto"

	"github.com/0chain/gosdk/zcnbridge"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zcnbridge/errors"
)

type ProofOfBurn struct {
	TxnID           string `json:"0chain_txn_id"`
	Nonce           int64  `json:"nonce"`
	Amount          int64  `json:"amount"`
	EthereumAddress string `json:"ethereum_address"`
	Signature       []byte `json:"signature,omitempty"`
}

func (pb *ProofOfBurn) Encode() []byte {
	return encryption.RawHash(pb)
}

func (pb *ProofOfBurn) Decode(input []byte) error {
	return json.Unmarshal(input, pb)
}

func (pb *ProofOfBurn) Verify() (err error) {
	switch {
	case pb.TxnID == "":
		err = errors.NewError("failed to verify proof of burn ticket", "0chain txn id is required")
	case pb.Nonce == 0:
		err = errors.NewError("failed to verify proof of burn ticket", "Nonce is required")
	case pb.Amount == 0:
		err = errors.NewError("failed to verify proof of burn ticket", "Amount is required")
	case pb.EthereumAddress == "":
		err = errors.NewError("failed to verify proof of burn ticket", "Receiving client id is required")
	}
	return
}

func (pb *ProofOfBurn) UnsignedMessage() string {
	return fmt.Sprintf("%v:%v:%v:%v", pb.TxnID, pb.Amount, pb.Nonce, pb.EthereumAddress)
}

func (pb *ProofOfBurn) SignWithEthereum(b *zcnbridge.BridgeClient) (err error) {
	sig, err := b.SignWithEthereumChain(pb.UnsignedMessage())
	if err != nil {
		return errors.Wrap("signature_ethereum", "failed to sign proof-of-burn ticket", err)
	}
	pb.Signature = sig

	return
}

// Sign can sign if chain config is initialized
func (pb *ProofOfBurn) Sign() (err error) {
	hash := zcncrypto.Sha3Sum256(pb.UnsignedMessage())
	sig, err := client.Sign(hash)
	if err != nil {
		return errors.Wrap("signature_0chain", "failed to sign proof-of-burn ticket using walletString ID ", err)
	}
	pb.Signature = []byte(sig)

	return
}

// SignWith0Chain can sign with the provided walletString
func (pb *ProofOfBurn) SignWith0Chain(w *zcncrypto.Wallet) (err error) {
	hash := zcncrypto.Sha3Sum256(pb.UnsignedMessage())
	sig, err := zcncore.SignWith0Wallet(hash, w)
	if err != nil {
		return errors.Wrap("signature_0chain", "failed to sign proof-of-burn ticket using walletString ID "+w.ClientID, err)
	}
	pb.Signature = []byte(sig)

	return
}
