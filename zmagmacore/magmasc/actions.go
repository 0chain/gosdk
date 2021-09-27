package magmasc

import (
	"context"
	"encoding/json"

	"github.com/0chain/gosdk/zmagmacore/transaction"
)

// ExecuteSessionStart starts session for provided IDs by executing ConsumerSessionStartFuncName.
func ExecuteSessionStart(ctx context.Context, sessID string) (*Acknowledgment, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	ackn, err := RequestAcknowledgment(sessID)
	if err != nil {
		return nil, err
	}
	input, err := json.Marshal(&ackn)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(
		ctx,
		Address,
		ConsumerSessionStartFuncName,
		string(input),
		ackn.Terms.GetAmount(),
	)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	ackn = new(Acknowledgment)
	if err := json.Unmarshal([]byte(txn.TransactionOutput), &ackn); err != nil {
		return nil, err
	}

	return ackn, err
}

// ExecuteDataUsage executes ProviderDataUsageFuncName and returns current Acknowledgment.
func ExecuteDataUsage(
	ctx context.Context, downloadBytes, uploadBytes uint64, sessID string, sessTime uint32) (*Acknowledgment, error) {

	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	dataUsage := DataUsage{
		DownloadBytes: downloadBytes,
		UploadBytes:   uploadBytes,
		SessionID:     sessID,
		SessionTime:   sessTime,
	}
	input, err := json.Marshal(&dataUsage)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(ctx, Address, ProviderDataUsageFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	ackn := Acknowledgment{}
	if err = json.Unmarshal([]byte(txn.TransactionOutput), &ackn); err != nil {
		return nil, err
	}

	return &ackn, err
}

// ExecuteSessionStop requests Acknowledgment from the blockchain and executes ConsumerSessionStopFuncName
// and verifies including the transaction in the blockchain.
//
// Returns Acknowledgment for session with provided ID.
func ExecuteSessionStop(ctx context.Context, sessionID string) (*Acknowledgment, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	// need to respond billing to compute value of txn
	ackn, err := RequestAcknowledgment(sessionID)
	if err != nil {
		return nil, err
	}

	input, err := json.Marshal(&ackn)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(
		ctx,
		Address,
		ConsumerSessionStopFuncName,
		string(input),
		ackn.Billing.Amount,
	)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	ackn = new(Acknowledgment)
	if err := json.Unmarshal([]byte(txn.TransactionOutput), ackn); err != nil {
		return nil, err
	}

	return ackn, err
}

// ExecuteProviderRegister executes provider registration magma sc function and returns current Provider.
func ExecuteProviderRegister(ctx context.Context, provider *Provider) (*Provider, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input, err := json.Marshal(provider)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(ctx, Address, ProviderRegisterFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	provider = &Provider{}
	if err = json.Unmarshal([]byte(txn.TransactionOutput), provider); err != nil {
		return nil, err
	}

	return provider, nil
}

// ExecuteProviderUpdate executes update provider magma sc function and returns updated Provider.
func ExecuteProviderUpdate(ctx context.Context, provider *Provider) (*Provider, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := provider.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, Address, ProviderUpdateFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	provider = new(Provider)
	if err := json.Unmarshal([]byte(txn.TransactionOutput), provider); err != nil {
		return nil, err
	}

	return provider, nil
}

// ExecuteConsumerRegister executes consumer registration magma sc function and returns current Consumer.
func ExecuteConsumerRegister(ctx context.Context, consumer *Consumer) (*Consumer, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input, err := json.Marshal(consumer)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(ctx, Address, ConsumerRegisterFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	consumer = &Consumer{}
	if err = json.Unmarshal([]byte(txn.TransactionOutput), consumer); err != nil {
		return nil, err
	}

	return consumer, nil
}

// ExecuteConsumerUpdate executes update consumer magma sc function and returns updated Consumer.
func ExecuteConsumerUpdate(ctx context.Context, consumer *Consumer) (*Consumer, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := consumer.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, Address, ConsumerUpdateFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	consumer = new(Consumer)
	if err := json.Unmarshal([]byte(txn.TransactionOutput), consumer); err != nil {
		return nil, err
	}

	return consumer, nil
}

// ExecuteSessionInit executes session init magma sc function and returns Acknowledgment.
func ExecuteSessionInit(ctx context.Context, consExtID, provExtID, apID, sessID string, terms ProviderTerms) (*Acknowledgment, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	ackn := Acknowledgment{
		Consumer: &Consumer{
			ExtID: consExtID,
		},
		Provider: &Provider{
			ExtID: provExtID,
		},
		AccessPointID: apID,
		SessionID:     sessID,
		Terms:         terms,
	}
	input, err := json.Marshal(&ackn)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(
		ctx,
		Address,
		ProviderSessionInitFuncName,
		string(input),
		0,
	)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	ackn = Acknowledgment{}
	if err := json.Unmarshal([]byte(txn.TransactionOutput), &ackn); err != nil {
		return nil, err
	}

	return &ackn, err
}
