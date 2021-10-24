// go:build test
// +build test

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"testing"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/wasmsdk/httpwasm"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/stretchr/testify/assert"
)

func TestWasmSDK(t *testing.T) {
	Logger.Info("Testing WASM SDK")

	sharder := httpwasm.NewSharderServer()
	defer sharder.Close()
	miner := httpwasm.NewMinerServer()
	defer miner.Close()

	var miner_dummy = js.Global().Call("eval", fmt.Sprintf("({minerString: %#v})", miner.URL+"/miner02"))

	var sharders_dummy = js.Global().Call("eval", fmt.Sprintf("({shardersArray: [%#v, %#v]})", sharder.URL+"/sharder02", sharder.URL+"/sharder03"))

	t.Run("Test SDK SetNetwork", func(t *testing.T) {
		setNetwork := js.FuncOf(SetNetwork)
		defer setNetwork.Release()

		miners := miner_dummy.Get("minerString")
		sharders := sharders_dummy.Get("shardersArray")

		assert.Empty(t, setNetwork.Invoke(miners, sharders).Truthy())
		assert.Equal(t, blockchain.GetMiners()[0], miner.URL+"/miner02")
		assert.Equal(t, blockchain.GetSharders()[0], sharder.URL+"/sharder02")
		assert.Equal(t, blockchain.GetSharders()[1], sharder.URL+"/sharder03")
	})

	t.Run("Test GetNetwork", func(t *testing.T) {
		setNetwork := js.FuncOf(SetNetwork)
		defer setNetwork.Release()

		jsMiner := js.Global().Call("eval", fmt.Sprintf("({minerString: %#v})", miner.URL+"/miner01"))
		jsSharder := js.Global().Call("eval", fmt.Sprintf("({sharderString: %#v})", sharder.URL+"/sharder01"))

		miners := jsMiner.Get("minerString")
		sharders := jsSharder.Get("sharderString")

		assert.Empty(t, setNetwork.Invoke(miners, sharders).Truthy())

		getNetwork := js.FuncOf(GetNetwork)
		defer getNetwork.Release()
		res := getNetwork.Invoke()

		assert.Equal(t, miner.URL+"/miner01", res.Get("miners").String())
		assert.Equal(t, sharder.URL+"/sharder01", res.Get("sharders").String())
	})

	t.Run("Test SetMaxTxnQuery", func(t *testing.T) {
		assert.Equal(t, 5, blockchain.GetMaxTxnQuery())

		setMaxTxnQuery := js.FuncOf(SetMaxTxnQuery)
		defer setMaxTxnQuery.Release()

		assert.Empty(t, setMaxTxnQuery.Invoke("1").Truthy())
		assert.Equal(t, 1, blockchain.GetMaxTxnQuery())
	})

	t.Run("Test SetQuerySleepTime", func(t *testing.T) {
		assert.Equal(t, 5, blockchain.GetQuerySleepTime())

		setQuerySleepTime := js.FuncOf(SetQuerySleepTime)
		defer setQuerySleepTime.Release()

		assert.Empty(t, setQuerySleepTime.Invoke("1").Truthy())
		assert.Equal(t, 1, blockchain.GetQuerySleepTime())
	})

	t.Run("Test SetMinSubmit", func(t *testing.T) {
		assert.Equal(t, 50, blockchain.GetMinSubmit())

		setMinSubmit := js.FuncOf(SetMinSubmit)
		defer setMinSubmit.Release()

		assert.Empty(t, setMinSubmit.Invoke("2").Truthy())
		assert.Equal(t, 2, blockchain.GetMinSubmit())
	})

	t.Run("Test SetMinConfirmation", func(t *testing.T) {
		assert.Equal(t, 50, blockchain.GetMinConfirmation())

		setMinConfirmation := js.FuncOf(SetMinConfirmation)
		defer setMinConfirmation.Release()

		assert.Empty(t, setMinConfirmation.Invoke("2").Truthy())
		assert.Equal(t, 2, blockchain.GetMinConfirmation())
	})

	t.Run("Test CreateReadPool", func(t *testing.T) {
		createReadPool := js.FuncOf(CreateReadPool)
		defer createReadPool.Release()
		result, err := await(createReadPool.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Get("result").IsNull())
	})

	t.Run("Test AllocFilter", func(t *testing.T) {
		mockAllocID := httpwasm.GetMockAllocationId(0)
		var allocPoolStat sdk.AllocationPoolStats

		jsPoolStats := js.Global().Call("eval", fmt.Sprintf(`({poolStats: {pools:[{id:%#v,balance:1000,expire_at:1641016719,allocation_id:%#v,blobbers:[],locked:false},{id:%#v,balance:100,expire_at:1641016719,allocation_id:%#v,blobbers:[],locked:true}],back:{id:%#v,balance:150}}})`, httpwasm.GetMockId(0), mockAllocID, httpwasm.GetMockId(100), httpwasm.GetMockAllocationId(100), httpwasm.GetMockId(150)))
		jsAllocID := js.Global().Call("eval", fmt.Sprintf(`({allocID: %#v})`, mockAllocID))

		poolStats := jsPoolStats.Get("poolStats")
		allocID := jsAllocID.Get("allocID")

		allocFilter := js.FuncOf(AllocFilter)
		defer allocFilter.Release()
		result := allocFilter.Invoke(poolStats, allocID).String()
		err := json.Unmarshal([]byte(result), &allocPoolStat)

		assert.Empty(t, err)
		assert.Equal(t, common.Key(mockAllocID), allocPoolStat.Pools[0].AllocationID)
	})

	t.Run("Test GetReadPoolInfo", func(t *testing.T) {
		getReadPoolInfo := js.FuncOf(GetReadPoolInfo)
		defer getReadPoolInfo.Release()

		jsClientID := js.Global().Call("eval", fmt.Sprintf(`({clientID: %#v})`, httpwasm.GetMockId(0)))
		clientID := jsClientID.Get("clientID")

		result, err := await(getReadPoolInfo.Invoke(clientID))

		aps := result[0].Get("result").String()

		var allocPoolStat sdk.AllocationPoolStats

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &allocPoolStat))
		assert.Equal(t, clientID.String(), allocPoolStat.Pools[0].ID)
	})

	t.Run("Test ReadPoolLock", func(t *testing.T) {
		readPoolLock := js.FuncOf(ReadPoolLock)
		defer readPoolLock.Release()

		jsReadPool := js.Global().Call("eval", fmt.Sprintf(`({duration: "60m",allocID: %#v,blobberID: %#v,tokens: 2500,fee: 150})`, httpwasm.GetMockAllocationId(1), httpwasm.GetMockBlobberId(1)))

		duration := jsReadPool.Get("duration")
		allocID := jsReadPool.Get("allocID")
		blobberID := jsReadPool.Get("blobberID")
		tokens := jsReadPool.Get("tokens")
		fee := jsReadPool.Get("fee")

		result, err := await(readPoolLock.Invoke(duration, allocID, blobberID, tokens, fee))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].IsNull())
	})

	t.Run("Test ReadPoolUnlock", func(t *testing.T) {
		readPoolUnlock := js.FuncOf(ReadPoolUnlock)
		defer readPoolUnlock.Release()

		jsReadPool := js.Global().Call("eval", fmt.Sprintf(`({poolID: %#v, fee: 150})`, httpwasm.GetMockId(1)))

		poolID := jsReadPool.Get("poolID")
		fee := jsReadPool.Get("fee")

		result, err := await(readPoolUnlock.Invoke(poolID, fee))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].IsNull())
	})

	t.Run("Test GetStakePoolInfo", func(t *testing.T) {
		getStakePoolInfo := js.FuncOf(GetStakePoolInfo)
		defer getStakePoolInfo.Release()

		jsBlobberID := js.Global().Call("eval", fmt.Sprintf(`({blobberID: %#v})`, httpwasm.GetMockId(0)))
		blobberID := jsBlobberID.Get("blobberID")

		result, err := await(getStakePoolInfo.Invoke(blobberID))

		aps := result[0].Get("result").String()

		var stakePoolInfo sdk.StakePoolInfo

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &stakePoolInfo))
		assert.Equal(t, common.Key(blobberID.String()), stakePoolInfo.ID)
	})

	t.Run("Test GetStakePoolUserInfo", func(t *testing.T) {
		getStakePoolUserInfo := js.FuncOf(GetStakePoolUserInfo)
		defer getStakePoolUserInfo.Release()

		jsClientID := js.Global().Call("eval", fmt.Sprintf(`({clientID: %#v})`, httpwasm.GetMockId(0)))
		clientID := jsClientID.Get("clientID")

		result, err := await(getStakePoolUserInfo.Invoke(clientID))

		aps := result[0].Get("result").String()

		var stakePoolUserInfo sdk.StakePoolUserInfo

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &stakePoolUserInfo))
		assert.Equal(t, common.Key(clientID.String()), stakePoolUserInfo.Pools[common.Key(clientID.String())][0].ID)
	})

	t.Run("Test StakePoolLock", func(t *testing.T) {
		stakePoolLock := js.FuncOf(StakePoolLock)
		defer stakePoolLock.Release()

		jsReadPool := js.Global().Call("eval", fmt.Sprintf(`({blobberID: %#v,value: 2500,fee: 150})`, httpwasm.GetMockBlobberId(1)))

		blobberID := jsReadPool.Get("blobberID")
		value := jsReadPool.Get("value")
		fee := jsReadPool.Get("fee")

		result, err := await(stakePoolLock.Invoke(blobberID, value, fee))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Get("result").Truthy())
	})

	t.Run("Test StakePoolUnlock", func(t *testing.T) {
		stakePoolUnlock := js.FuncOf(StakePoolUnlock)
		defer stakePoolUnlock.Release()

		jsReadPool := js.Global().Call("eval", fmt.Sprintf(`({blobberID:%#v, poolID: %#v, fee: 150})`, httpwasm.GetMockBlobberId(1), httpwasm.GetMockId(1)))

		blobberID := jsReadPool.Get("blobberID")
		poolID := jsReadPool.Get("poolID")
		fee := jsReadPool.Get("fee")

		result, err := await(stakePoolUnlock.Invoke(blobberID, poolID, fee))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, common.Timestamp(1641016719).ToTime().Format(time.RFC850), result[0].Get("result").String())
	})

	t.Run("Test StakePoolPayInterests", func(t *testing.T) {
		stakePoolPayInterests := js.FuncOf(StakePoolPayInterests)
		defer stakePoolPayInterests.Release()

		jsBlobberID := js.Global().Call("eval", fmt.Sprintf(`({blobberID: %#v})`, httpwasm.GetMockId(0)))
		blobberID := jsBlobberID.Get("blobberID")

		result, err := await(stakePoolPayInterests.Invoke(blobberID))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].IsNull())
	})

	t.Run("Test GetWritePoolInfo", func(t *testing.T) {
		getWritePoolInfo := js.FuncOf(GetWritePoolInfo)
		defer getWritePoolInfo.Release()

		jsClientID := js.Global().Call("eval", fmt.Sprintf(`({clientID: %#v})`, httpwasm.GetMockId(0)))
		clientID := jsClientID.Get("clientID")

		result, err := await(getWritePoolInfo.Invoke(clientID))

		aps := result[0].Get("result").String()

		var allocPoolStat sdk.AllocationPoolStats

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &allocPoolStat))
		assert.Equal(t, clientID.String(), allocPoolStat.Pools[0].ID)
	})

	t.Run("Test WritePoolLock", func(t *testing.T) {
		writePoolLock := js.FuncOf(WritePoolLock)
		defer writePoolLock.Release()

		jsWritePool := js.Global().Call("eval", fmt.Sprintf(`({duration: "60m",allocID: %#v,blobberID: %#v,tokens: 2500,fee: 150})`, httpwasm.GetMockAllocationId(1), httpwasm.GetMockBlobberId(1)))

		duration := jsWritePool.Get("duration")
		allocID := jsWritePool.Get("allocID")
		blobberID := jsWritePool.Get("blobberID")
		tokens := jsWritePool.Get("tokens")
		fee := jsWritePool.Get("fee")

		result, err := await(writePoolLock.Invoke(duration, allocID, blobberID, tokens, fee))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].IsNull())
	})

	t.Run("Test WritePoolUnlock", func(t *testing.T) {
		writePoolUnlock := js.FuncOf(WritePoolUnlock)
		defer writePoolUnlock.Release()

		jsWritePool := js.Global().Call("eval", fmt.Sprintf(`({poolID: %#v, fee: 150})`, httpwasm.GetMockId(1)))

		poolID := jsWritePool.Get("poolID")
		fee := jsWritePool.Get("fee")

		result, err := await(writePoolUnlock.Invoke(poolID, fee))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].IsNull())
	})

	t.Run("Test GetChallengePoolInfo", func(t *testing.T) {
		getChallengePoolInfo := js.FuncOf(GetChallengePoolInfo)
		defer getChallengePoolInfo.Release()

		jsAllocID := js.Global().Call("eval", fmt.Sprintf(`({allocID: %#v})`, httpwasm.GetMockAllocationId(0)))
		allocID := jsAllocID.Get("allocID")

		result, err := await(getChallengePoolInfo.Invoke(allocID))

		aps := result[0].Get("result").String()

		var challengePool sdk.ChallengePoolInfo

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &challengePool))
		assert.Equal(t, allocID.String(), challengePool.ID)
	})

	t.Run("Test GetStorageSCConfig", func(t *testing.T) {
		getStorageSCConfig := js.FuncOf(GetStorageSCConfig)
		defer getStorageSCConfig.Release()

		result, err := await(getStorageSCConfig.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.NotEmpty(t, true, result[0].String())
	})

	t.Run("Test GetBlobbers", func(t *testing.T) {
		getBlobbers := js.FuncOf(GetBlobbers)
		defer getBlobbers.Release()

		result, err := await(getBlobbers.Invoke())

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Get("result").Truthy())
	})

	t.Run("Test GetBlobber", func(t *testing.T) {
		getBlobber := js.FuncOf(GetBlobber)
		defer getBlobber.Release()

		jsBlobberID := js.Global().Call("eval", fmt.Sprintf(`({blobberID: %#v})`, httpwasm.GetMockBlobberId(1)))
		blobberID := jsBlobberID.Get("blobberID")

		result, err := await(getBlobber.Invoke(blobberID))

		aps := result[0].String()

		var blobber sdk.Blobber

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &blobber))
		assert.Equal(t, common.Key(blobberID.String()), blobber.ID)
		assert.Equal(t, common.Timestamp(1633878133), blobber.LastHealthCheck)
	})

	t.Run("Test GetClientEncryptedPublicKey", func(t *testing.T) {
		getClientEncryptedPublicKey := js.FuncOf(GetClientEncryptedPublicKey)
		defer getClientEncryptedPublicKey.Release()

		result, err := await(getClientEncryptedPublicKey.Invoke())
		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, "XTNjdLsxHO5+gU6WG9J8au7dvy406FZtQF3DResVx/E=", result[0].String())
	})

	t.Run("Test GetAllocation", func(t *testing.T) {
		getAllocation := js.FuncOf(GetAllocation)
		defer getAllocation.Release()

		jsAllocID := js.Global().Call("eval", fmt.Sprintf(`({allocID: %#v})`, httpwasm.GetMockAllocationId(0)))
		allocID := jsAllocID.Get("allocID")

		result, err := await(getAllocation.Invoke(allocID))

		aps := result[0].String()

		var allocation sdk.Allocation

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &allocation))
		assert.Equal(t, allocID.String(), allocation.ID)
	})

	t.Run("Test SetNumBlockDownloads", func(t *testing.T) {
		setNumBlockDownloads := js.FuncOf(SetNumBlockDownloads)
		defer setNumBlockDownloads.Release()

		jsNumBlock := js.Global().Call("eval", "({numBlock: 100})")
		numBlock := jsNumBlock.Get("numBlock")

		assert.Equal(t, true, setNumBlockDownloads.Invoke(numBlock).IsNull())

	})

	t.Run("Test GetAllocationsForClient", func(t *testing.T) {
		getAllocationsForClient := js.FuncOf(GetAllocationsForClient)
		defer getAllocationsForClient.Release()

		jsClientID := js.Global().Call("eval", fmt.Sprintf(`({clientID: %#v})`, httpwasm.GetMockId(0)))
		clientID := jsClientID.Get("clientID")

		result, err := await(getAllocationsForClient.Invoke(clientID))

		aps := result[0].String()

		var allocation []*sdk.Allocation

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &allocation))
		assert.Equal(t, clientID.String(), allocation[0].ID)
	})

	t.Run("Test CreateAllocationForOwner", func(t *testing.T) {
		createAllocationForOwner := js.FuncOf(CreateAllocationForOwner)
		defer createAllocationForOwner.Release()

		readPrice := &sdk.PriceRange{
			Min: 200,
			Max: 1000,
		}
		readPriceJSON, _ := json.Marshal(readPrice)
		writePrice := &sdk.PriceRange{
			Min: 200,
			Max: 1000,
		}
		writePriceJSON, _ := json.Marshal(writePrice)

		jsAllocOwner := js.Global().Call("eval", fmt.Sprintf(`({owner: %#v, ownerpublickey: %#v, datashards: 2000, parityshards: 1000, size: 500, expiry: 1633878133, readPrice: %#v, writePrice: %#v, mcct: "60h", lock: 500, preferredBlobbers: []})`, client.GetClientID(), client.GetClientPublicKey(), string(readPriceJSON), string(writePriceJSON)))

		owner := jsAllocOwner.Get("owner")
		ownerpublickey := jsAllocOwner.Get("ownerpublickey")
		datashards := jsAllocOwner.Get("datashards")
		parityshards := jsAllocOwner.Get("parityshards")
		size := jsAllocOwner.Get("size")
		expiry := jsAllocOwner.Get("expiry")
		readPriceArgs := jsAllocOwner.Get("readPrice")
		writePriceArgs := jsAllocOwner.Get("writePrice")
		mcct := jsAllocOwner.Get("mcct")
		lock := jsAllocOwner.Get("lock")
		preferredBlobbers := jsAllocOwner.Get("preferredBlobbers")

		result, err := await(createAllocationForOwner.Invoke(owner, ownerpublickey, datashards, parityshards, size, expiry, readPriceArgs, writePriceArgs, mcct, lock, preferredBlobbers))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Truthy())
	})

	t.Run("Test UpdateAllocation", func(t *testing.T) {
		updateAllocation := js.FuncOf(UpdateAllocation)
		defer updateAllocation.Release()

		jsAllocOwner := js.Global().Call("eval", fmt.Sprintf(`({size: 1000000, expiry: 1633878133, allocID: %#v, lock: 500, immutable: true})`, httpwasm.GetMockAllocationId(1)))

		size := jsAllocOwner.Get("size")
		expiry := jsAllocOwner.Get("expiry")
		allocID := jsAllocOwner.Get("allocID")
		lock := jsAllocOwner.Get("lock")
		immutable := jsAllocOwner.Get("immutable")

		result, err := await(updateAllocation.Invoke(size, expiry, allocID, lock, immutable))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Truthy())
	})

	t.Run("Test FinalizeAllocation", func(t *testing.T) {
		finalizeAllocation := js.FuncOf(FinalizeAllocation)
		defer finalizeAllocation.Release()

		jsFinalizeAlloc := js.Global().Call("eval", fmt.Sprintf(`({allocID: %#v})`, httpwasm.GetMockAllocationId(1)))

		allocID := jsFinalizeAlloc.Get("allocID")

		result, err := await(finalizeAllocation.Invoke(allocID))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Truthy())
	})

	t.Run("Test CancelAllocation", func(t *testing.T) {
		cancelAllocation := js.FuncOf(CancelAllocation)
		defer cancelAllocation.Release()

		jsCancelAlloc := js.Global().Call("eval", fmt.Sprintf(`({allocID: %#v})`, httpwasm.GetMockAllocationId(1)))

		allocID := jsCancelAlloc.Get("allocID")

		result, err := await(cancelAllocation.Invoke(allocID))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Truthy())
	})

	t.Run("Test UpdateBlobberSettings", func(t *testing.T) {
		updateBlobberSettings := js.FuncOf(UpdateBlobberSettings)
		defer updateBlobberSettings.Release()

		blob := &sdk.Blobber{
			ID:              common.Key(httpwasm.GetMockBlobberId(1)),
			Capacity:        common.Size(1000000),
			Used:            common.Size(500000),
			LastHealthCheck: common.Timestamp(1633878133),
		}

		blobJSON, _ := json.Marshal(blob)
		jsBlob := js.Global().Call("eval", fmt.Sprintf(`({blob: %#v})`, string(blobJSON)))
		blobArgs := jsBlob.Get("blob")

		result, err := await(updateBlobberSettings.Invoke(blobArgs))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, true, result[0].Truthy())
	})

	t.Run("Test CommitToFabric", func(t *testing.T) {
		commitToFabric := js.FuncOf(CommitToFabric)
		defer commitToFabric.Release()

		server := httpwasm.NewDefaultServer()
		defer server.Close()

		var fabricMockConfig struct {
			URL  string `json:"url"`
			Body struct {
				Channel          string   `json:"channel"`
				ChaincodeName    string   `json:"chaincode_name"`
				ChaincodeVersion string   `json:"chaincode_version"`
				Method           string   `json:"method"`
				Args             []string `json:"args"`
			} `json:"body"`
			Auth struct {
				Username string `json:"username"`
				Password string `json:"password"`
			} `json:"auth"`
		}

		fabricMockConfig.URL = server.URL + "/commitfabric"
		fabricMockConfig.Body.Channel = httpwasm.GetMockId(15)
		fabricMockConfig.Body.ChaincodeName = httpwasm.GetMockId(200)
		fabricMockConfig.Body.ChaincodeVersion = "0.0.1"
		fabricMockConfig.Body.Method = "GET"
		fabricMockConfig.Body.Args = []string{}
		fabricMockConfig.Auth.Username = "TEST"
		fabricMockConfig.Auth.Password = "TEST"

		fabricJSON, _ := json.Marshal(fabricMockConfig)
		fmt.Println(string(fabricJSON))
		jsFabric := js.Global().Call("eval", fmt.Sprintf(`({fabric: %#v, metaTxnData: "TEST"})`, string(fabricJSON)))
		fabricArgs := jsFabric.Get("fabric")
		metaTxnData := jsFabric.Get("metaTxnData")

		result, err := await(commitToFabric.Invoke(metaTxnData, fabricArgs))
		aps := result[0].String()

		var fabricMockResponse struct {
			Channel          string   `json:"channel"`
			ChaincodeName    string   `json:"chaincode_name"`
			ChaincodeVersion string   `json:"chaincode_version"`
			Method           string   `json:"method"`
			Args             []string `json:"args"`
		}

		assert.Equal(t, true, err[0].IsNull())
		assert.Empty(t, json.Unmarshal([]byte(aps), &fabricMockResponse))
		assert.Equal(t, fabricMockConfig.Body.Channel, fabricMockResponse.Channel)
		assert.Equal(t, fabricMockConfig.Body.ChaincodeName, fabricMockResponse.ChaincodeName)
	})

	t.Run("Test GetAllocationMinLock", func(t *testing.T) {
		getAllocationMinLock := js.FuncOf(GetAllocationMinLock)
		defer getAllocationMinLock.Release()

		readPrice := &sdk.PriceRange{
			Min: 200,
			Max: 1000,
		}
		readPriceJSON, _ := json.Marshal(readPrice)
		writePrice := &sdk.PriceRange{
			Min: 200,
			Max: 1000,
		}
		writePriceJSON, _ := json.Marshal(writePrice)

		jsAllocOwner := js.Global().Call("eval", fmt.Sprintf(`({datashards: 2000, parityshards: 1000, size: 500, expiry: 1633878133, readPrice: %#v, writePrice: %#v, mcct: "60h"})`, string(readPriceJSON), string(writePriceJSON)))

		datashards := jsAllocOwner.Get("datashards")
		parityshards := jsAllocOwner.Get("parityshards")
		size := jsAllocOwner.Get("size")
		expiry := jsAllocOwner.Get("expiry")
		readPriceArgs := jsAllocOwner.Get("readPrice")
		writePriceArgs := jsAllocOwner.Get("writePrice")
		mcct := jsAllocOwner.Get("mcct")

		result, err := await(getAllocationMinLock.Invoke(datashards, parityshards, size, expiry, readPriceArgs, writePriceArgs, mcct))

		assert.Equal(t, true, err[0].IsNull())
		assert.Equal(t, expiry.Int(), result[0].Int())
	})
}
