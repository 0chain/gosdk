# 0chain/wasmsdk
The 0chain wasm SDK is written in Go programming language, and released with WebAssembly binary format 

*NOTE* please use `try{await zcn.sdk.[method];}catch(err){...}` to handle any error from wasm sdk in js

## ZCN methods

### zcn.sdk.init
init wasm sdk 
  Input:
  > chainID, blockWorker, signatureScheme string, minConfirmation, minSubmit, confirmationChainLength int

  Output:
  > N/A


### zcn.jsProxy.setWallet 
set bls.SecretKey on runtime env(browser,nodejs...etc), and call `zcn.sdk.setWallet` to set wallet on go.

**Input**:
> bls, clientID, sk, pk string

**Output**:
> N/A

### zcn.sdk.setWallet
set wallet on go

**Input**:
> clientID,publicKey string

**Output**:
> N/A

### zcn.sdk.getEncryptedPublicKey
get encrypted public key by mnemonic
**Input**:
> mnemonic string

**Output**:
> string



## Blobber methods
### zcn.sdk.delete
delete remote file from blobbers

**Input**:
> allocationID, remotePath string,commit bool

**Output**:
> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)

**Example**:
```json
{
   "hash":"0da2f2ffb64e16629752626866c44855c9038e8459b83f6b913b86444809a6e2",
   "version":"1.0",
   "client_id":"bec04d9120f56ef4198ad0b75b09e34dbcebd79d77807ff4badf2094c5198090",
   "public_key":"92e88784e6cd8dd2f5328177757704112daa0368f28d599bf76825b5a98fbb02c796358dfe566efeacb96a1108f8851b1b4763d06db44c715e8ac80867322000",
   "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
   "transaction_data":"{\"CrudType\":\"Delete\",\"MetaData\":{\"Name\":\"scan4.png\",\"Type\":\"f\",\"Path\":\"/scan4.png\",\"LookupHash\":\"507a75dfb031dc3e888be1ffdbd51bb3b520fd5b4df46dbaa660040f8d3494ed\",\"Hash\":\"adab389e89121db0ab94a2b2137a28647851bde2827304a779784017b7c3dca5\",\"MimeType\":\"image/png\",\"Size\":14554,\"ActualFileSize\":14554,\"ActualNumBlocks\":1,\"EncryptedKey\":\"\",\"CommitMetaTxns\":[{\"ref_id\":66,\"txn_id\":\"c81c4772a9ce9e5a1f1c2398ea696be26e3b0e92658920593a79f96489afe395\",\"created_at\":\"2021-12-09T02":"18":15.767812Z\"}],"Collaborators":[],\"Attributes\":{}}}",
   "transaction_value":0,
   "signature":"313fc544caebd89deb2f1b89506cdef39c739b7068c86f399009db6b98eee184",
   "creation_date":1639016421,
   "transaction_type":10,
   "transaction_fee":0,
   "txn_output_hash":"a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
   "transaction_status":1
}
```

### zcn.sdk.rename
rename a file existing already on dStorage. Only the allocation's owner can rename a file.

**Input**:
> allocationID, remotePath, destName string, commit bool

**Output**:
> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)

### zcn.sdk.copy
copy file to another folder path on blobbers
**Input**:
> allocationID, remotePath, destPath string, commit bool


**Output**:
> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)

### zcn.sdk.move
move file to another remote folder path on dStorage. Only the owner of the allocation can copy an object.

**Input**:
> allocationID, remotePath, destPath string, commit bool

**Output**:
> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)


### zcn.sdk.share
generate an authtoken that provides authorization to the holder to the specified file on the remotepath.

**Input**:
> allocationID, remotePath, clientID, encryptionPublicKey string, expiration int, revoke bool,availableAfter int

**Output**:
> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)

### zcn.sdk.download
download your own or a shared file.

**Input**:
> allocationID, remotePath, authTicket, lookupHash string, downloadThumbnailOnly, autoCommit, rxPay, live, delay bool, blocksPerMarker, startBlock, endBlock int

**Output**:
>  {fileName:string, txn:transaction.Transaction, error:string}

**Example**
```json
{
   "fileName":"scan3.png",
   "url":"blob:http://localhost:3000/42157751-1d33-4448-88c8-7d7e2ad887a5",
   "txn":{
      "hash":"b43a2f20b29538ece7aa78103517bafa89804605a4f51519b20c6131427b627f",
      "version":"1.0",
      "client_id":"bec04d9120f56ef4198ad0b75b09e34dbcebd79d77807ff4badf2094c5198090",
      "public_key":"92e88784e6cd8dd2f5328177757704112daa0368f28d599bf76825b5a98fbb02c796358dfe566efeacb96a1108f8851b1b4763d06db44c715e8ac80867322000",
      "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
      "transaction_data":"{\\""CrudType\\"":\\""Download\\"",\\""MetaData\\"":{\\""Name\\"":\\"scan3.png\\",\\""Type\\"":\\""f\\"",\\""Path\\"":\\"/scan3.png\\",\\""LookupHash\\"":\\"71babab33e8f8e6719fd840b6f3e0d97aa279133c968e3333e04cd29ac2b1686\\",\\""Hash\\"":\\"adab389e89121db0ab94a2b2137a28647851bde2827304a779784017b7c3dca5\\",\\""MimeType\\"":\\""image/png\\"",\\""Size\\"":14554,\\""ActualFileSize\\"":14554,\\""ActualNumBlocks\\"":1,\\""EncryptedKey\\"":\\""\\"",\\""CommitMetaTxns\\"":[],\\""Collaborators\\"":[],\\""Attributes\\"":{}}}",
      "transaction_value":0,
      "signature":"1f2c71beb252d27a48627f6678ff63802422a90b651ac00a7a0a46afafe71f95",
      "creation_date":1639017104,
      "transaction_type":10,
      "transaction_fee":0,
      "txn_output_hash":"a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
      "transaction_status":1,
      "owner_id":"bec04d9120f56ef4198ad0b75b09e34dbcebd79d77807ff4badf2094c5198090",
      "status":"COMPLETED"
   }
}

```

### zcn.sdk.upload
upload file(s)

**Input**:
allocationID, remotePath string, fileBytes, thumbnailBytes []byte, encrypt, commit bool, attrWhoPaysForReads string, isLiveUpload, isSyncUpload bool, chunkSize int, isUpdate, isRepair bool

**Output**:
> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)
  
