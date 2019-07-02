# 0CHAIN REST ENDPOINTS
--- 

0Chain rest endpoints can be broadly divided into Wallet Transactions endpoints and Storage Transactions endpoints. Depending on what endpoints you are using, you may want to choose the right SDK. Storage Transactions endpoints are currently supported only on 0Chain's *gosdk* while Wallet Transactions endpoints are supported on both *gosdk* and *js-client-sdk*.

**Note1:** Not all rest endpoints are supported on all of 0Chain nodes. Some are supported on Miners, and some are on Sharders, while others on Blobbers. The endpoints documentation below clearly mentions that.


## ENDPOINT: `/v1/client/put`

**Purpose** To register a wallet on Blockchain

**METHOD** POST

**Send To** Miners

**Input**

"id": wallet.ClientID

"public_key": wallet.ClientKey

**Output** 

Newly created wallet information 

***Sample Output***
```
"{\"client_id\":\"701fa94a02141e33c8fee526852a4bfa54ba4cb230af3ee5b3885a804956e941\",\"client_key\":\"81a864b6e0059074a1adfcfd71eae4cab154f9dee906ed81eba858882709b673\",\"keys\":[{\"public_key\":\"81a864b6e0059074a1adfcfd71eae4cab154f9dee906ed81eba858882709b673\",\"private_key\":\"5126ac5a56e65c7010d72b71054564620ace408ac1d2ea66929392f422d449bb81a864b6e0059074a1adfcfd71eae4cab154f9dee906ed81eba858882709b673\"}],\"mnemonics\":\"drip include antique differ what gentle where bicycle junior crime outer dilemma member fine drip series train certain black abuse female direct grant alcohol\",\"version\":\"1.0\",\"date_created\":\"2019-06-26 20:07:54.4614548 +0000 UTC m=+0.041726101\"}"
```
---

## ENDPOINT: `/v1/transaction/put`

**Purpose** To create a transaction on Blockchain

**METHOD** POST

**Send To** Miners

**Input** 

Transaction with below details 

``` go
type Transaction struct {
	Hash              string `json:"hash,omitempty"`
	Version           string `json:"version,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	ToClientID        string `json:"to_client_id,omitempty"`
	ChainID           string `json:"chain_id,omitempty"`
	TransactionData   string `json:"transaction_data,omitempty"`
	Value             int64  `json:"transaction_value,omitempty"`
	Signature         string `json:"signature,omitempty"`
	CreationDate      int64  `json:"creation_date,omitempty"`
	TransactionType   int    `json:"transaction_type,omitempty"`
	TransactionOutput string `json:"transaction_output,omitempty"`
	OutputHash        string `json:"txn_output_hash"`
}
```

**Output**

HTTP response status codes

Returned Transaction Details sent along with the transaction hash on the blockchain.


***Sample Input***
``` 
 {a8986181e09f01813ee6226c4605cc34654000d1bb70f20bcca0d590cb8511cb 1.0 721de8aeb895f9b6404ad3e3b0ea38e2cf5e27b304a310bfb9bd990c27d804e3 1926b6c84f89b50da73cca6ed9984f091ed5a084e7baf041d4fcb823f70ef15b 6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7 0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe {"name":"new_allocation_request","input":{"data_shards":3,"expiration_date":1569363680,"parity_shards":1,"size":20480}} 0 1919efe33850df82ea2be23b0de69742a036f90f4ebfa2a4c7fa53e3aa5d0af68345687fb76bc79a22c7310b67f91849abf505bb18120a38182118ff3a23dd0b 1561587680 1000  }
 ```

***Sample Output***
``` 
[{http://vira.devb.testnet-0chain.net:7071/v1/transaction/put 200 200 OK {"async":true,"entity":{"hash":"a8986181e09f01813ee6226c4605cc34654000d1bb70f20bcca0d590cb8511cb","version":"1.0","client_id":"721de8aeb895f9b6404ad3e3b0ea38e2cf5e27b304a310bfb9bd990c27d804e3","to_client_id":"6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7","chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe","transaction_data":"{\"name\":\"new_allocation_request\",\"input\":{\"data_shards\":3,\"expiration_date\":1569363680,\"parity_shards\":1,\"size\":20480}}","transaction_value":0,"signature":"1919efe33850df82ea2be23b0de69742a036f90f4ebfa2a4c7fa53e3aa5d0af68345687fb76bc79a22c7310b67f91849abf505bb18120a38182118ff3a23dd0b","creation_date":1561587680,"transaction_fee":0,"transaction_type":1000,"txn_output_hash":"","transaction_status":0}}]
```
---

## ENDPOINT: `/v1/transaction/get/confirmation`

**Purpose** To search for a transaction on the blockchain

**METHOD** GET

**Send To** Sharders

**Input** 

Transaction Hash of the interested transaction

**Output** 

On success: Transaction details on the blockchain

On failure: Error

***Sample Input***
``` 
http://cala.devb.testnet-0chain.net:7171/v1/transaction/get/confirmation?hash=bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632
```

***Sample Success Output*** 
``` 
{"version":"1.0","hash":"bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632","block_hash":"bbcb7670b955221167dcd90d79ba186b5ed08edebba0e3ccdc69a52c4f524996","txn":{"hash":"bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632","version":"1.0","client_id":"721de8aeb895f9b6404ad3e3b0ea38e2cf5e27b304a310bfb9bd990c27d804e3","to_client_id":"6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7","chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe","transaction_data":"{\"name\":\"new_allocation_request\",\"input\":{\"data_shards\":3,\"expiration_date\":1569363772,\"parity_shards\":1,\"size\":20480}}","transaction_value":0,"signature":"cdcf63b52eb35eed6efb81f50093b5e28419d8c5e6fc997b1cbeef05d79509d3f828d6351f7b44c9de3afa57b1f24d039c9e4f13f686d42a349a6b76d7adaf0a","creation_date":1561587772,"transaction_fee":0,"transaction_type":1000,"transaction_output":"{\"id\":\"bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632\",\"data_shards\":3,\"parity_shards\":1,\"size\":20480,\"expiration_date\":1569363772,\"blobbers\":[{\"id\":\"3de9840655ddded756b1e0a4229d71cf09ecdc9b90af2d76ec1752075628b251\",\"url\":\"http://vira.devb.testnet-0chain.net:5051\"},{\"id\":\"72d94629908fe904c2a0e3ef00435efe8e6826e6b2a5b4072600298aa9da6d20\",\"url\":\"http://virb.devb.testnet-0chain.net:5051\"},{\"id\":\"76509b2f5b411900bca1461960d66be316acd219f4a6de8287a54c88db153bde\",\"url\":\"http://calb.devb.testnet-0chain.net:5051\"},{\"id\":\"a60b743a7196ce75d4faf991d128ff97eeb6f62aa320c522dd0afd927312fef9\",\"url\":\"http://cala.devb.testnet-0chain.net:5051\"}],\"owner_id\":\"721de8aeb895f9b6404ad3e3b0ea38e2cf5e27b304a310bfb9bd990c27d804e3\",\"owner_public_key\":\"1926b6c84f89b50da73cca6ed9984f091ed5a084e7baf041d4fcb823f70ef15b\",\"stats\":{\"used_size\":0,\"num_of_writes\":0,\"num_of_reads\":0,\"total_challenges\":0,\"num_open_challenges\":0,\"num_success_challenges\":0,\"num_failed_challenges\":0,\"latest_closed_challenge\":\"\"},\"blobber_details\":[{\"blobber_id\":\"3de9840655ddded756b1e0a4229d71cf09ecdc9b90af2d76ec1752075628b251\",\"allocation_id\":\"bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632\",\"size\":5120,\"allocation_root\":\"\",\"write_marker\":null,\"stats\":{\"used_size\":0,\"num_of_writes\":0,\"num_of_reads\":0,\"total_challenges\":0,\"num_open_challenges\":0,\"num_success_challenges\":0,\"num_failed_challenges\":0,\"latest_closed_challenge\":\"\"}},{\"blobber_id\":\"72d94629908fe904c2a0e3ef00435efe8e6826e6b2a5b4072600298aa9da6d20\",\"allocation_id\":\"bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632\",\"size\":5120,\"allocation_root\":\"\",\"write_marker\":null,\"stats\":{\"used_size\":0,\"num_of_writes\":0,\"num_of_reads\":0,\"total_challenges\":0,\"num_open_challenges\":0,\"num_success_challenges\":0,\"num_failed_challenges\":0,\"latest_closed_challenge\":\"\"}},{\"blobber_id\":\"76509b2f5b411900bca1461960d66be316acd219f4a6de8287a54c88db153bde\",\"allocation_id\":\"bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632\",\"size\":5120,\"allocation_root\":\"\",\"write_marker\":null,\"stats\":{\"used_size\":0,\"num_of_writes\":0,\"num_of_reads\":0,\"total_challenges\":0,\"num_open_challenges\":0,\"num_success_challenges\":0,\"num_failed_challenges\":0,\"latest_closed_challenge\":\"\"}},{\"blobber_id\":\"a60b743a7196ce75d4faf991d128ff97eeb6f62aa320c522dd0afd927312fef9\",\"allocation_id\":\"bee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632\",\"size\":5120,\"allocation_root\":\"\",\"write_marker\":null,\"stats\":{\"used_size\":0,\"num_of_writes\":0,\"num_of_reads\":0,\"total_challenges\":0,\"num_open_challenges\":0,\"num_success_challenges\":0,\"num_failed_challenges\":0,\"latest_closed_challenge\":\"\"}}]}","txn_output_hash":"99f5d1354b0859f1ec683cc215f1d32bd4d858cb6dc9c84d79daf74a95409068","transaction_status":1},"creation_date":1561587773,"miner_id":"347495f9f2915f3205772cf444320f6cf1a8d0c7cd1da9d621c116d8bcb18e90","round":504863,"transaction_status":1,"round_random_seed":-55809941478188169,"merkle_tree_root":"b827a913c5637d68cd556e3a0be901c1fbe9cba1c3517cc619a17dc707339d6a","merkle_tree_path":{"nodes":["fa6b4cc136ec3438b1aefaef6b0173154c3544a55fe58cf5256c5cadcadf552a","562605f404b7b6d6e9481f598c13c445c73508f51fa29b3a295e0de60ee74f07","1823199b9203f566a551020728bdec12764772a00ee32a6866cbb5e58ca36e65","af182db90bda4912bc5fb38d45c22b01b7fd2fe4a7b6b88bf2c25f85f14e4036"],"leaf_index":0},"receipt_merkle_tree_root":"f3647bfc07074ece2715ff2d317f1e9d7fcc51214126d42f07e68b032927a826","receipt_merkle_tree_path":{"nodes":["a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a","c23f89dc588f71b1181366a014eae26b051d35ebea620cf0cd2a8466a758d8e1","9d12ea745f03257d4b1d6cc53ea2075d18e47f03c9f80c134cb8082253abc07f","cde5a1a83098b9438617129a187e69f1c1c6e493adf24a8649fb476b1cc27b4b"],"leaf_index":0}}
```

***Sample Failure Output***
```
{"code":"entity_not_found","error":"entity_not_found: txn_summary not found with id = eee436b5e9753cf6c1170202bb874bd539cdb3319740828c23efa62360c40632"}
```
---

## ENDPOINT: `/v1/client/get/balance`

**Purpose** To query balance of a wallet

**METHOD** GET

**Send To** Miners

**Input** 

Wallet's client_id

**Output** 

transaction hash of the pour transaction, 

round when the transaciton is processed,

wallet balance

***Sample Input***
``` 
http://cala.devb.testnet-0chain.net:7071/v1/client/get/balance?client_id=701fa94a02141e33c8fee526852a4bfa54ba4cb230af3ee5b3885a804956e941
```

***Sample Success Output***
{"txn":"a7621db1018234699888113a19377a6bafeee6a178366c31112c5946eb1941f8","round":150255,"balance":10000000000}

***Sample Failure Output***
{"error":"value not present"}

---

## ENDPOINT: `/v1/scstate/get`

**METHOD** GET

**Purpose** To get the lock token configuration information such as interest rate from blockchain.

**Send To** Sharders

**Input** 

Address of Interest Pool SmartContract

Key of Interest Pool SmartContract

Currently, both of them are predefined. See sample input for more details.

**Output** 

lock token configuration information

***Sample Input***
```
http://cala.devb.testnet-0chain.net:7171/v1/scstate/get?sc_address=6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9&key=6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9
```
---

## ENDPOINT:  `/v1/screst/` + InterestPoolSmartContractAddress + `/getPoolsStats`

Note: Currently InterestPoolSmartContractAddress is predefined. See sample input for more details.

**Purpose** To get the ealier locked token pool stats

**METHOD** GET

**Send To** Sharder

**Input** 

Interest Pool SmartContract Address, 

Wallet ID of the wallet that owns the locked tokens

**Output** locked token pool stats or a failure message

***Sample Input***
```
http://peda.devb.testnet-0chain.net:7171/v1/screst/6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d9/getPoolsStats?client_id=a406e19a179a5bd10b7d3dd5377f699f76967ad10a31d4805a3909e088e3e501
```
***Sample Failed Output***
```
"failed to get stats: no pools exist"
```

---

## ENDPOINT: "/v1/file/upload/"

**Purpose** To send a upload request

**METHOD** POST'

**Send To** Blobbers

**Input** allocation id, File Reader

***Sample Input***
```
{POST http://vira.devb.testnet-0chain.net:5051/v1/file/upload/ffb8114204029cc556e0d3f99b3c40a1c0d79f1f75240ef3c9da29f8ff474622 HTTP/1.1 1 1 map[X-App-Client-Id:[6202c8db2657205ce28515ba7c0e65b5acbd55aca4eae3feded3d4d6e8693743] X-App-Client-Key:[245859ccdb88628180de02482aeaac8c56d6ee1035d151a6119593fdaff7f7f4]] 0xc00000e360 <nil> 0 [] false vira.devb.testnet-0chain.net:5051 map[] map[] <nil> map[]   <nil> <nil> <nil> <nil>}
```
---

## ENDPOINT: "/v1/file/download/"

**Purpose** To download a file from remote

**METHOD** GET

**Send To** Blobbers

**Input** Allocation ID, File Reader

***Sample Input***
```
GET http://virb.devb.testnet-0chain.net:5051/v1/file/referencepath/adf87db53288d0f97ed2ec7db96db1d95376981eeefe4bee8e688e2270e234ab?path=%2FV%2FJ%2FC%2FJ%2FhYzRy.txt HTTP/1.1 1 1 map[X-App-Client-Id:[6202c8db2657205ce28515ba7c0e65b5acbd55aca4eae3feded3d4d6e8693743] X-App-Client-Key:[245859ccdb88628180de02482aeaac8c56d6ee1035d151a6119593fdaff7f7f4]] <nil> <nil> 0 [] false virb.devb.testnet-0chain.net:5051 map[] map[] <nil> map[]   <nil> <nil> <nil> <nil>}
```
---

## ENDPOINT: '/v1/file/list/'

**Purpose** To get folder structure and file details from blobbers

**METHOD** GET

**Send To** Blobbers

**Input** Allocation ID, remote path 

**Output** folders and files at the remote path

***Sample Input***
```
GET http://cala.devb.testnet-0chain.net:5051/v1/file/list/adf87db53288d0f97ed2ec7db96db1d95376981eeefe4bee8e688e2270e234ab?path=%2F HTTP/1.1 1 1 map[] <nil> <nil> 0 [] false cala.devb.testnet-0chain.net:5051 map[] map[] <nil> map[]   <nil> <nil> <nil> <nil>
```

---

## ENDPOINT: '/v1/file/referencepath/'

**Purpose** To get reference path for a given file

**METHOD** GET

**Send To** Blobbers

**Input** 
Allocation ID, 

absolute remote path

***Sample Input***
```
GET http://virb.devb.testnet-0chain.net:5051/v1/file/referencepath/adf87db53288d0f97ed2ec7db96db1d95376981eeefe4bee8e688e2270e234ab?path=%2FV%2FJ%2FC%2FJ%2FhYzRy.txt HTTP/1.1 1 1 map[X-App-Client-Id:[6202c8db2657205ce28515ba7c0e65b5acbd55aca4eae3feded3d4d6e8693743] X-App-Client-Key:[245859ccdb88628180de02482aeaac8c56d6ee1035d151a6119593fdaff7f7f4]] <nil> <nil> 0 [] false virb.devb.testnet-0chain.net:5051 map[] map[] <nil> map[]   <nil> <nil> <nil> <nil>}
```
***Sample Output***
```
[Reference path: {"meta_data":{"hash":"","lookup_hash":"","name":"/","num_of_blocks":0,"path":"/","path_hash":"","type":"d"},"latest_write_marker":null}
```
---

## ENDPOINT: "/v1/connection/commit/" 

**Purpose** To commit the upload transaction

**METHOD** POST

**Send To** BLOBBERS

**Input** allocation id

***Sample Input***
```
http://calb.devb.testnet-0chain.net:5051/v1/connection/commit/c471ab76b0bb2766b29cd97830d310d9f45d3a8c0e9bee979a029af37d9acbe3
```
---

## ENDPOINT: "/v1/file/meta/"

**Purpose** To get meta data of a file

**METHOD** POST

**Send To** Blobbers

**Input** allocation id

***Sample Input***
```
POST http://cala.devb.testnet-0chain.net:5051/v1/file/meta/c471ab76b0bb2766b29cd97830d310d9f45d3a8c0e9bee979a029af37d9acbe3
```
---
