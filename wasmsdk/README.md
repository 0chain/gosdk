# 0chain/wasmsdk
The 0chain wasm SDK is written in Go programming language, and released with WebAssembly binary format 

*NOTE* please use `try{await zcn.sdk.[method];}catch(err){...}` to handle any error from wasm sdk in js

## ZCN methods

- `zcn.sdk.init(chainID, blockWorker, signatureScheme string, minConfirmation, minSubmit, confirmationChainLength int)`: init wasm sdk 
- `zcn.bls.setWallet(bls,clientID, sk, pk string)`: set bls.SecretKey on js, and call `zcn.sdk.setWallet` to set wallet on go.
- `zcn.sdk.setWallet(clientID,publicKey string)`: set wallet on go


## Blobber methods
- `zcn.sdk.delete(allocationID, remotePath string,commit bool)`: delete remote file from blobbers

  