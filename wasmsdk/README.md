# 0chain/wasmsdk
The 0chain wasm SDK is written in Go programming language, and released with WebAssembly binary format

*NOTE* please use `try{await zcn.sdk.[method];}catch(err){...}` to handle any error from wasm sdk in js

## ZCN methods

### zcn.sdk.init
init wasm sdk

**Input**:
  > chainID, blockWorker, signatureScheme string, minConfirmation, minSubmit, confirmationChainLength int,zboxHost, zboxAppType string

**Output**:
  > N/A

### zcn.sdk.hideLogs
hide interactive sdk logs. default is hidden.

**Input**:
  > N/A

**Output**:
  > N/A

### zcn.sdk.showLogs
show interactive sdk logs. default is hidden.

**Input**:
  > N/A

**Output**:
  > N/A


### zcn.sdk.getUSDRate
get USD rate by token symbol(eg zcn, eth)

**Input**:
  > symbol string

**Output**:
  > float64

### zcn.sdk.isWalletID
valid wallet id

**Input**:
  > clientID string

**Output**:
  > bool

### zcn.jsProxy.setWallet
set bls.SecretKey on runtime env(browser,nodejs...etc), and call `zcn.sdk.setWallet` to set wallet on go.

**Input**:
> bls, clientID, sk, pk string

**Output**:
> N/A


### zcn.sdk.setWallet
set wallet on go

**Input**:
> clientID, publicKey, privateKey string

**Output**:
> N/A


**Input**:
> host string

**Output**:
> N/A

### zcn.sdk.getPublicEncryptionKey
get public encryption key by mnemonic

**Input**:
> mnemonic string

**Output**:
> string


### zcn.sdk.getAllocationBlobbers
get blobbers with filters for creating allocation

**Input**:
> referredBlobberURLs []string,
	dataShards, parityShards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64

**Output**:
> string array

### zcn.sdk.createAllocation
create an allocation

**Input**:
> name string, datashards, parityshards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64,preferredBlobberIds []string

**Output**:
> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)


### zcn.sdk.listAllocations
list all allocations

**Input**:
> N/A

**Output**:
> [sdk.Allocation](https://github.com/0chain/gosdk/blob/a9e504e4a0e8fc76a05679e4ef183bb03b8db8e5/zboxcore/sdk/allocation.go#L140) array

### zcn.sdk.transferAllocation
changes the owner of an allocation. Only a curator or the current owner of the allocation, can change an allocation's ownership.

**Input**:
> allocationId, newOwnerId, newOwnerPublicKey string

**Output**:
> N/A


### zcn.sdk.freezeAllocation
freeze allocation so that data can no longer be modified

**Input**:
> allocationId string

**Output**:
> N/A




### zcn.sdk.getWalletBalance
get wallet balance

**Input**:
> clientId string

**Output**:
> {zcn:float64, usd: float64}

### zcn.sdk.getBlobberIds
convert blobber urls to blobber ids

**Input**:
> blobberUrls []string

**Output**:
> []string


### zcn.sdk.createReadPool
create readpool in storage SC if the pool is missing.

**Input**:
> N/A

**Output**:
> string


## Blobber methods
### zcn.sdk.delete
delete remote file from blobbers

**Input**:
> allocationID, remotePath string

**Output**:
> {commandSuccess:bool,error:string}



### zcn.sdk.rename
rename a file existing already on dStorage. Only the allocation's owner can rename a file.

**Input**:
> allocationID, remotePath, destName string

**Output**:
> {commandSuccess:bool, error:string}

### zcn.sdk.copy
copy file to another folder path on blobbers
**Input**:
> allocationID, remotePath, destPath string


**Output**:
> {commandSuccess:bool, error:string}

### zcn.sdk.move
move file to another remote folder path on dStorage. Only the owner of the allocation can copy an object.

**Input**:
> allocationID, remotePath, destPath string

**Output**:
> {commandSuccess:bool, error:string}


### zcn.sdk.share
generate an authtoken that provides authorization to the holder to the specified file on the remotepath.

**Input**:
> allocationID, remotePath, clientID, encryptionPublicKey string, expiration int, revoke bool,availableAfter int

**Output**:
> string

### zcn.sdk.download
download your own or a shared file.

**Input**:
> allocationID, remotePath, authTicket, lookupHash string, downloadThumbnailOnly bool, numBlocks int

**Output**:
>  {commandSuccess:bool, fileName:string,url:string, error:string}

**Example**
```json
{
   "commandSuccess":true,
   "fileName":"scan3.png",
   "url":"blob:http://localhost:3000/42157751-1d33-4448-88c8-7d7e2ad887a5",
}

```

### zcn.sdk.downloadBlocks
download blocks of a file

**Input**:
> allocationID, remotePath, authTicket, lookupHash string, numBlocks int, startBlockNumber, endBlockNumber int64

**Output**:
>  {commandSuccess:bool, fileName:string,url:string, error:string}

**Example**
```json
{
   "commandSuccess":true,
   "fileName":"scan3.png",
   "url":"blob:http://localhost:3000/42157751-1d33-4448-88c8-7d7e2ad887a5",
}

```

### zcn.sdk.upload
upload file(s)

**Input**:
> allocationID, remotePath string, fileBytes, thumbnailBytes []byte, encrypt bool, isUpdate, isRepair bool, numBlocks int

**Output**:
> {commandSuccess:bool, error:string}


### zcn.sdk.play
play stream video files

**Input**:
> allocationID, remotePath, authTicket, lookupHash string, isLive bool

**Output**:
> N/A


### zcn.sdk.stop
stop current play

**Input**:
> N/A

**Output**:
> N/A


### zcn.sdk.listObjects
list files with allocationID and remotePath

**Input**:
> allocationId string, remotePath string

**Output**:
> sdk.ListResult


### zcn.sdk.createDir
create folder from blobbers

**Input**:
> allocationID, remotePath string

**Output**:
> N/A

### zcn.sdk.getFileStats

**Input**:
> allocationID string, remotePath string

**Output**:
> string: []sdk.FileStats


## Swap methods
### zcn.sdk.setSwapWallets

**Input**:
> usdcTokenAddress, bancorAddress, zcnTokenAddress, ethWalletMnemonic string

**Output**:
> N/A

### zcn.sdk.swapToken

**Input**:
> swapAmount int64, tokenSource string

**Output**:
> string: txnHash



## 0Box API methods
### zcn.sdk.getCsrfToken
get a fresh CSRF token

**Input**:
> N/A

**Output**:
> string

### zcn.sdk.createJwtSession
create a jwt session with phone number
**Input**:
> phoneNumber string

**Output**:
> sessionID int64

### zcn.sdk.createJwtToken
create a jwt token with jwt session id and otp
**Input**:
> phoneNumber string, jwtSessionID int64, otp string

**Output**:
> token string


### zcn.sdk.refreshJwtToken
refresh jwt token
**Input**:
> phoneNumber string, token string

**Output**:
> token string

