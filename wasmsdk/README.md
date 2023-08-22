# 0chain/wasmsdk

The 0chain wasm SDK is written in Go programming language, and released with WebAssembly binary format

_NOTE_ please use `try{await zcn.sdk.[method];}catch(err){...}` to handle any error from wasm sdk in js

## ZCN global js methods

### zcn.setWallet

set bls.SecretKey on runtime env(browser,nodejs...etc), and call `zcn.sdk.setWallet` to set wallet on go.

**Input**:

> bls, clientID, sk, pk, mnemonic string

**Output**:

> N/A

### zcn.bulkUpload

bulk upload files. it will wrap options, and call `zcn.sdk.bulkUpload` to process upload

**Input**:

> bulkOptions: [ { allocationId:string,remotePath:string,file:FileReader, thumbnailBytes:[]byte, encrypt:bool,isUpdate:bool,isRepair:bool,numBlocks:int,callback:function(totalBytes, completedBytes, error) } ]

**Output**:

> [ {remotePath:"/d.png", success:true,error:""} ]

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

### zcn.sdk.createThumbnail

create thumbnail of an image buffer

**Input**:

> buf []byte, width,height int

**Output**:

> []byte

### zcn.sdk.isWalletID

valid wallet id

**Input**:

> clientID string

**Output**:

> bool

### zcn.sdk.setWallet

set wallet on go

**Input**:

> clientID, publicKey, privateKey string

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

> datashards, parityshards int, size,

    minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64,preferredBlobberIds []string, setThirdPartyExtendable  bool

**Output**:

> [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)

### zcn.sdk.listAllocations

list all allocations

**Input**:

> N/A

**Output**:

> [sdk.Allocation](https://github.com/0chain/gosdk/blob/a9e504e4a0e8fc76a05679e4ef183bb03b8db8e5/zboxcore/sdk/allocation.go#L140) array

### zcn.sdk.getAllocation

get allocation detail

**Input**:

> allocationID string

**Output**:

> [sdk.Allocation](https://github.com/0chain/gosdk/blob/a9e504e4a0e8fc76a05679e4ef183bb03b8db8e5/zboxcore/sdk/allocation.go#L140)

### zcn.sdk.reloadAllocation

clean cache, and get allocation detail from blockchain

**Input**:

> allocationID string

**Output**:

> [sdk.Allocation](https://github.com/0chain/gosdk/blob/a9e504e4a0e8fc76a05679e4ef183bb03b8db8e5/zboxcore/sdk/allocation.go#L140)

### zcn.sdk.transferAllocation

changes the owner of an allocation. Only the current owner of the allocation, can change an allocation's ownership.

**Input**:

> allocationId, newOwnerId, newOwnerPublicKey string

**Output**:

> N/A

### zcn.sdk.freezeAllocation

freeze allocation so that data can no longer be modified

**Input**:

> allocationId string

**Output**:

> hash: string

### zcn.sdk.cancelAllocation

immediately return all remaining tokens from challenge pool back to the allocation's owner and cancels the allocation. If blobbers already got some tokens, the tokens will not be returned. Remaining min lock payment to the blobber will be funded from the allocation's write pools.

**Input**:

> allocationId string

**Output**:

> hash: string

### zcn.sdk.updateAllocation

updates allocation settings

**Input**:

> allocationId string, name string,size int64, extend bool,lock int64,setImmutable, updateTerms bool,addBlobberId, removeBlobberId string, setThirdPartyExtendable  bool

**Output**:

> hash: string

### zcn.sdk.getAllocationWith

returns allocation from authToken

**Input**:

> authTicket string

**Output**:

> [sdk.Allocation](https://github.com/0chain/gosdk/blob/a9e504e4a0e8fc76a05679e4ef183bb03b8db8e5/zboxcore/sdk/allocation.go#L140)

### zcn.sdk.getReadPoolInfo

gets information about the read pool for the allocation

**Input**:

> clientID string

**Output**:

> [sdk.ReadPool](https://github.com/0chain/gosdk/blob/6878504e4e4d7cb25b2ac819c3c578228b3d3e30/zboxcore/sdk/sdk.go#L167-L169)

### zcn.sdk.decodeAuthTicket

**Input**:

> ticket string

**Output**:

> recipientPublicKey string, markerStr string, tokensInSAS uint64

### zcn.sdk.lockWritePool

locks given number of tokes for given duration in write pool

**Input**:

> allocationId string, tokens uint64, fee uint64

**Output**:

> hash: string

### zcn.sdk.lockStakePool

locks given number of tokens on a provider

**Input**:

> [providerType](https://github.com/0chain/gosdk/blob/bc96f54e68a68ef5d757428b9c4153405ebe4163/zboxcore/sdk/sdk.go#L1186-L1194) uint64, tokens uint64, fee uint64, providerID string,

**Output**:

> hash: string

### zcn.sdk.unlockStakePool

unlocks tokens on a provider

**Input**:

> [providerType](https://github.com/0chain/gosdk/blob/bc96f54e68a68ef5d757428b9c4153405ebe4163/zboxcore/sdk/sdk.go#L1186-L1194) uint64, fee uint64, providerID string,

**Output**:

> returns time where the tokens can be unlocked

### zcn.sdk.getSkatePoolInfo

get the details of the stakepool associated with provider
**Input**:

> [providerType](https://github.com/0chain/gosdk/blob/bc96f54e68a68ef5d757428b9c4153405ebe4163/zboxcore/sdk/sdk.go#L1186-L1194) int, providerID string

**Output**:

> [sdk.StakePoolInfo](https://github.com/0chain/gosdk/blob/2ec97a9bb116db166e31c0207971282e7008d22c/zboxcore/sdk/sdk.go#L263-L275), err

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

### zcn.sdk.getBlobbers

get blobbers from the network

**Input**:

**Output**:

> array of [sdk.Blobber](https://github.com/0chain/gosdk/blob/6878504e4e4d7cb25b2ac819c3c578228b3d3e30/zboxcore/sdk/sdk.go#L558-L572)

### zcn.sdk.createReadPool

create readpool in storage SC if the pool is missing.

**Input**:

> N/A

**Output**:

> string

### zcn.sdk.executeSmartContract

send raw SmartContract transaction, and verify result

**Input**:

> address, methodName, input string, value uint64

**Output**:

> > [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)

### zcn.sdk.faucet

faucet SmartContract

**Input**:

> methodName, input string, token float64

**Output**:

> > [transaction.Transaction](https://github.com/0chain/gosdk/blob/e1e35e084d5c17d6bf233bbe8ac9c91701bdd8fd/core/transaction/entity.go#L32)

## Blobber methods

### zcn.sdk.getLookupHash

get lookup hash by allocation id and path

**Input**:

> allocationID string, path string

**Output**:

> string

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

> allocationID, remotePath, authTicket, lookupHash string, downloadThumbnailOnly bool, numBlocks int,callbackFuncName string

**Output**:

> {commandSuccess:bool, fileName:string,url:string, error:string}

**Example**

```json
{
  "commandSuccess": true,
  "fileName": "scan3.png",
  "url": "blob:http://localhost:3000/42157751-1d33-4448-88c8-7d7e2ad887a5"
}
```

### zcn.sdk.downloadBlocks

download blocks of a file

**Input**:

> allocationID, remotePath, authTicket, lookupHash string, numBlocks int, startBlockNumber, endBlockNumber int64, callbackFuncName string

**Output**:

> {commandSuccess:bool, fileName:string,url:string, error:string}

**Example**

```json
{
  "commandSuccess": true,
  "fileName": "scan3.png",
  "url": "blob:http://localhost:3000/42157751-1d33-4448-88c8-7d7e2ad887a5"
}
```

### zcn.sdk.upload

upload file

**Input**:

> allocationID, remotePath string, fileBytes, thumbnailBytes []byte, encrypt bool, isUpdate, isRepair bool, numBlocks int

**Output**:

> {commandSuccess:bool, error:string}

### zcn.sdk.bulkUpload

bulk upload files with json options

**Input**:

> jsonBulkUploadOptions string:
> BulkOption: { allocationId,remotePath,readChunkFuncName, fileSize, thumbnailBytes, encrypt,isUpdate,isRepair,numBlocks,callbackFuncName:callbackFuncName }

**Output**:

> [ {remotePath:"/d.png", success:true,error:""} ]

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

### zcn.sdk.updateBlobberSettings

**Input**:

> blobberSettingsJson string
> blobberSettings: fetch blobber settings by calling /getBlobber on sharder and modify them (see demo for usage)
> **Output**:
> string: resp

### zcn.sdk.getAllocationMinLock

**Input**:

> datashards int, parityshards int, size int, maxreadPrice int, maxwritePrice int

**Output**:

> int: min_lock_demand


### zcn.sdk.getUpdateAllocationMinLock

**Input**:

> allocationID string, size int, extend bool, updateTerms bool, addBlobberId string, removeBlobberId string

**Output**:

> int: min_lock_demand

### zcn.sdk.getRemoteFileMap

takes allocation ID and returns all the files/directories in allocation as JSON
**Input**:

> allocationID string

**Output**:

> []\*fileResp

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
