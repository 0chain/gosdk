# Bridge SDK documentation

This bridge SDK comprises 3 types of client and following available methods

1. Bridge deployer/owner *for* `Authorizer Smart Contract`
2. Customer/client *to execute* `Bridge Smart contract` burn and mint transactions
3. Token bridge authorizer *for* signing payload for 0Chain and Ethereum chain

## Bridge owner SDK

**Deployer or Owner SDK** is required to Add or Remove Authorizers from `Authorizers Smart Contract`

Configuration file: ```~/.zcn/owner.yaml``` to initialize owner

```yaml
owner:
    # Address of Ethereum authorizers contract
    AuthorizersAddress: "0xFE20Ce9fBe514397427d20C91CB657a4478A0FFa"
    OwnerEthereumMnemonic: "add here the owner mnemonic that has deployed auth and bridge contracts"
```

Authorizes config: `~/.zcn/authorizers.yaml` to add authorizers

**Example:**

```go
transaction, err := owner.AddEthereumAuthorizer(context.TODO(), wallet.Address)
if err != nil {
    fmt.Println(err)
}

status, err := zcnbridge.ConfirmEthereumTransaction(transaction.Hash().String(), 5, time.Second*5)
if err != nil {
    fmt.Println(err)
}

```

## Client SDK

This SDK is used to execute mint and burn transaction on both chains, Ethereum and 0Chain

Configuration file: ```~/.zcn/bridge.yaml``` to initialize the client

Functionality:

- Burn (0Chain and Ethereum)
- Mint (0Chain and Ethereum)

For more detailed information please find documentation [here](https://github.com/0chain/0chain/blob/staging/code/go/0chain.net/smartcontract/zcnsc/README.MD)

## Token Bridge SDK

Functionality:

1. Used by token bridge to sign payloads for 0Chain and Ethereum chains
2. Checks transactions status and give the client payload to
