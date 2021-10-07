package magmasc

import (
	"context"
	"encoding/json"

	"github.com/0chain/gosdk/zmagmacore/magmasc/pb"
	"github.com/0chain/gosdk/zmagmacore/transaction"
)

// ExecuteSessionStart starts session for provided IDs by executing ConsumerSessionStartFuncName.
func ExecuteSessionStart(ctx context.Context, sessID string) (*Session, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	billRatio, err := FetchBillingRatio()
	if err != nil {
		return nil, err
	}

	session, err := RequestSession(sessID)
	if err != nil {
		return nil, err
	}
	input, err := json.Marshal(&session)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(
		ctx,
		Address,
		ConsumerSessionStartFuncName,
		string(input),
		session.AccessPoint.Terms.GetAmount()*billRatio,
	)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	session = new(Session)
	if err := json.Unmarshal([]byte(txn.TransactionOutput), &session); err != nil {
		return nil, err
	}

	return session, err
}

// ExecuteDataUsage executes ProviderDataUsageFuncName and returns current Session.
func ExecuteDataUsage(ctx context.Context, marker *DataMarker) (*Session, error) {

	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := marker.Encode()
	txnHash, err := txn.ExecuteSmartContract(ctx, Address, ProviderDataUsageFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	session := Session{}
	if err = json.Unmarshal([]byte(txn.TransactionOutput), &session); err != nil {
		return nil, err
	}

	return &session, err
}

// ExecuteSessionStop requests Session from the blockchain and executes ConsumerSessionStopFuncName
// and verifies including the transaction in the blockchain.
//
// Returns Session for session with provided ID.
func ExecuteSessionStop(ctx context.Context, sessionID string) (*Session, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	// need to respond billing to compute value of txn
	session, err := RequestSession(sessionID)
	if err != nil {
		return nil, err
	}

	input, err := json.Marshal(&session)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(
		ctx,
		Address,
		ConsumerSessionStopFuncName,
		string(input),
		session.Billing.Amount,
	)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	session = new(Session)
	if err := json.Unmarshal([]byte(txn.TransactionOutput), session); err != nil {
		return nil, err
	}

	return session, err
}

// ExecuteProviderRegister executes provider registration magma sc function and returns current Provider.
func ExecuteProviderRegister(ctx context.Context, provider *Provider) (*Provider, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := provider.Encode()
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
	if err = provider.Decode([]byte(txn.TransactionOutput)); err != nil {
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

	provider = &Provider{}
	if err = provider.Decode([]byte(txn.TransactionOutput)); err != nil {
		return nil, err
	}

	return provider, nil
}

// ExecuteProviderStake stakes the provider tokens of the magma sc and returns Provider.
func ExecuteProviderStake(ctx context.Context, provider *Provider) (*Provider, error) {
	minStake, err := ProviderMinStakeFetch()
	if err != nil {
		return nil, err
	}

	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := provider.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, Address, ProviderStakeFuncName, string(input), minStake)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	provider = &Provider{}
	if err = provider.Decode([]byte(txn.TransactionOutput)); err != nil {
		return nil, err
	}

	return provider, nil
}

// ExecuteProviderUnStake unstaked provider tokens of the magma sc and returns Provider.
func ExecuteProviderUnStake(ctx context.Context, provider *Provider) (*Provider, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := provider.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, Address, ProviderUnStakeFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	provider = &Provider{}
	if err = provider.Decode([]byte(txn.TransactionOutput)); err != nil {
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

// ExecuteAccessPointRegister executes access point registration magma sc function and returns current AccessPoint.
func ExecuteAccessPointRegister(ctx context.Context, accessPoint *AccessPoint) (*AccessPoint, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input, err := json.Marshal(accessPoint)
	if err != nil {
		return nil, err
	}
	txnHash, err := txn.ExecuteSmartContract(ctx, Address, AccessPointRegisterFuncName, string(input), accessPoint.MinStake)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	accessPoint = &AccessPoint{}
	if err = json.Unmarshal([]byte(txn.TransactionOutput), accessPoint); err != nil {
		return nil, err
	}

	return accessPoint, nil
}

// ExecuteAccessPointUpdate executes update access point magma sc function and returns updated AccessPoint.
func ExecuteAccessPointUpdate(ctx context.Context, accessPoint *AccessPoint) (*AccessPoint, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := accessPoint.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, Address, AccessPointUpdateFuncName, string(input), accessPoint.MinStake)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	accessPoint = new(AccessPoint)
	if err := json.Unmarshal([]byte(txn.TransactionOutput), accessPoint); err != nil {
		return nil, err
	}

	return accessPoint, nil
}

// ExecuteSessionInit executes session init magma sc function and returns Session.
func ExecuteSessionInit(ctx context.Context, consExtID, provExtID, apID, sessID string) (*Session, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	session := Session{
		Consumer: &Consumer{
			ExtID: consExtID,
		},
		Provider: &Provider{
			Provider: &pb.Provider{
				ExtID: provExtID,
			},
		},
		SessionID: sessID,
		AccessPoint: &AccessPoint{
			ID: apID,
		},
	}
	input, err := json.Marshal(&session)
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

	session = Session{}
	if err := json.Unmarshal([]byte(txn.TransactionOutput), &session); err != nil {
		return nil, err
	}

	return &session, err
}

// ExecuteUserRegister executes user registration magma sc function and returns current User.
func ExecuteUserRegister(ctx context.Context, user *User) (*User, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := user.Encode()
	txnHash, err := txn.ExecuteSmartContract(ctx, Address, UserRegisterFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, txnHash)
	if err != nil {
		return nil, err
	}

	user = &User{}
	if err = user.Decode([]byte(txn.TransactionOutput)); err != nil {
		return nil, err
	}

	return user, nil
}

// ExecuteUserUpdate executes update user magma sc function and returns updated User.
func ExecuteUserUpdate(ctx context.Context, user *User) (*User, error) {
	txn, err := transaction.NewTransactionEntity()
	if err != nil {
		return nil, err
	}

	input := user.Encode()
	hash, err := txn.ExecuteSmartContract(ctx, Address, UserUpdateFuncName, string(input), 0)
	if err != nil {
		return nil, err
	}

	txn, err = transaction.VerifyTransaction(ctx, hash)
	if err != nil {
		return nil, err
	}

	user = &User{}
	if err = user.Decode([]byte(txn.TransactionOutput)); err != nil {
		return nil, err
	}

	return user, nil
}
