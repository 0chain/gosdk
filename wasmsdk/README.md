# 0chain/wasmsdk
The 0chain wasm SDK is written in Go programming language, and released with WebAssembly binary format 

*NOTE* please use `try{await zcn.sdk.[method];}catch(err){...}` to handle any error from wasm sdk in js

## ZCN methods

### `zcn.sdk.init`: init wasm sdk 
  Input:
  > chainID, blockWorker, signatureScheme string, minConfirmation, minSubmit, confirmationChainLength int

  Output:
  > empty


- `zcn.bls.setWallet(bls,clientID, sk, pk string)`: set bls.SecretKey on js, and call `zcn.sdk.setWallet` to set wallet on go.
- `zcn.sdk.setWallet(clientID,publicKey string)`: set wallet on go
- `zcn.sdk.getEncryptedPublicKey(mnemonic string)`: get encrypted public key by mnemonic


## Blobber methods
- `zcn.sdk.delete(allocationID, remotePath string,commit bool)`:    delete remote file from blobbers
- `zcn.sdk.rename(allocationID, remotePath, destName string, commit bool)`: rename a file existing already on dStorage. Only the allocation's owner can rename a file.
- `zcn.sdk.copy(allocationID, remotePath, destPath string, commit bool)`:   copy file to another folder path on blobbers
- `zcn.sdk.move(allocationID, remotePath, destPath string, commit bool)`:   move file to another remote folder path on dStorage. Only the owner of the allocation can copy an object.
- `zcn.sdk.share(allocationID, remotePath, clientID, encryptionPublicKey string, expiration int, revoke bool,availableAfter int)`:    generate an authtoken that provides authorization to the holder to the specified file on the remotepath.
- `zcn.sdk.download(allocationID, remotePath, authTicket, lookupHash string, downloadThumbnailOnly, autoCommit, rxPay, live, delay bool, blocksPerMarker, startBlock, endBlock int)`: download your own or a shared file.


  