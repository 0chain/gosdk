package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	// "sync"
	"syscall/js"

	// "github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// convert JS String to []String
func ZBOXstrToListSring(s string) []string {
	slice := []string{}
	err := json.Unmarshal([]byte(s), &slice)

	if err != nil {
		panic(err)
	}
	return slice
}

func strToPriceRange(s string) sdk.PriceRange {
	var p sdk.PriceRange
	err := json.Unmarshal([]byte(s), &p)
	if err == nil {
		fmt.Println("error:", err)
	}

	return p
}

func strToBlob(s string) sdk.Blobber {
	var b sdk.Blobber
	err := json.Unmarshal([]byte(s), &b)
	if err == nil {
		fmt.Println("error:", err)
	}

	return b
}

func InitAuthTicket(this js.Value, p []js.Value) interface{} {
	authTicket := p[0].String()
	result = sdk.InitAuthTicket(authTicket)
	return result
}

func ZBOXSetLogLevel(this js.Value, p []js.Value) interface{} {
	logLevel, _ := strconv.Atoi(p[0].String())

	sdk.SetLogLevel(logLevel)
	return nil
}

func ZBOXSetLogFile(this js.Value, p []js.Value) interface{} {
	logFile := p[0].String()
	verbose, _ := strconv.ParseBool(p[1].String())

	sdk.SetLogFile(logFile, verbose)
	return nil
}

func GetNetwork(this js.Value, p []js.Value) interface{} {
	result := sdk.GetNetwork()
	return result
}

func SetMaxTxnQuery(this js.Value, p []js.Value) interface{} {
	num, _ := strconv.Atoi(p[0].String())
	sdk.SetMaxTxnQuery(num)
	return nil
}

func SetQuerySleepTime(this js.Value, p []js.Value) interface{} {
	num, _ := strconv.Atoi(p[0].String())
	sdk.SetQuerySleepTime(num)
	return nil
}

func SetMinSubmit(this js.Value, p []js.Value) interface{} {
	num, _ := strconv.Atoi(p[0].String())
	sdk.SetMinSubmit(num)
	return nil
}

func SetMinConfirmation(this js.Value, p []js.Value) interface{} {
	num, _ := strconv.Atoi(p[0].String())
	sdk.SetMinConfirmation(num)
	return nil
}

func ZBOXSetNetwork(this js.Value, p []js.Value) interface{} {
	miners := ZBOXstrToListSring(p[0].String())
	sharders := ZBOXstrToListSring(p[1].String())
	sdk.SetNetwork(miners, sharders)
	return nil
}

// //
// // read pool
// //

func CreateReadPool(this js.Value, p []js.Value) interface{} {
	err := sdk.CreateReadPool()
	if err != nil {
		fmt.Println("Cannot create read pool")
	}
	return err
}

func AllocFilter(this js.Value, p []js.Value) interface{} {
	poolStats := p[0].String()
	allocID := p[1].String()

	var alloc sdk.AllocationPoolStats
	err := json.Unmarshal([]byte(poolStats), &alloc)
	if err == nil {
		fmt.Println("error:", err)
	}
	allocFilter := (*sdk.AllocationPoolStats).AllocFilter

	allocFilter(&alloc, allocID)
	return nil
}

func ZBOXGetReadPoolInfo(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()
	result, err := sdk.GetReadPoolInfo(clientID)
	if err != nil {
		return err
	}
	return result
}

// // ReadPoolLock locks given number of tokes for given duration in read pool.
func ReadPoolLock(this js.Value, p []js.Value) interface{} {
	dur, _ := time.ParseDuration(p[0].String()) // time.Duration,
	allocID := p[1].String()
	blobberID := p[2].String()
	tokens, _ := strconv.ParseInt(p[3].String(), 10, 64)
	fee, _ := strconv.ParseInt(p[4].String(), 10, 64)

	err := sdk.ReadPoolLock(dur, allocID, blobberID, tokens, fee)
	if err != nil {
		fmt.Println("Cannot set wallet info")
	}
	return err
}

// // ReadPoolUnlock unlocks tokens in expired read pool
func ReadPoolUnlock(this js.Value, p []js.Value) interface{} {
	poolID := p[0].String()
	fee, _ := strconv.ParseInt(p[1].String(), 10, 64)

	err := sdk.ReadPoolUnlock(poolID, fee)
	if err != nil {
		fmt.Println("Cannot set wallet info")
	}
	return err
}

// //
// // stake pool
// //

// // GetStakePoolInfo for given client, or, if the given clientID is empty,
// // for current client of the sdk.
func ZBOXGetStakePoolInfo(this js.Value, p []js.Value) interface{} {
	blobberID := p[0].String()
	result, err := sdk.GetStakePoolInfo(blobberID)
	if err != nil {
		return err
	}
	return result
}

// // GetStakePoolUserInfo obtains blobbers/validators delegate pools statistic
// // for a user. If given clientID is empty string, then current client used.
func ZBOXGetStakePoolUserInfo(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()
	result, err := sdk.GetStakePoolUserInfo(clientID)
	if err != nil {
		return err
	}
	return result
}

// // StakePoolLock locks tokens lack in stake pool
func StakePoolLock(this js.Value, p []js.Value) interface{} {
	blobberID := p[0].String()
	value, _ := strconv.ParseInt(p[3].String(), 10, 64)
	fee, _ := strconv.ParseInt(p[4].String(), 10, 64)

	result, err := sdk.StakePoolLock(blobberID, value, fee)
	if err != nil {
		return err
	}
	return result
}

// // StakePoolUnlock unlocks a stake pool tokens. If tokens can't be unlocked due
// // to opened offers, then it returns time where the tokens can be unlocked,
// // marking the pool as 'want to unlock' to avoid its usage in offers in the
// // future. The time is maximal time that can be lesser in some cases. To
// // unlock tokens can't be unlocked now, wait the time and unlock them (call
// // this function again).
func StakePoolUnlock(this js.Value, p []js.Value) interface{} {
	blobberID := p[0].String()
	poolID := p[1].String()
	fee, _ := strconv.ParseInt(p[2].String(), 10, 64)

	result, err := sdk.StakePoolUnlock(blobberID, poolID, fee)
	if err != nil {
		return err
	}
	return result
}

// // StakePoolPayInterests unlocks a stake pool rewards.
func StakePoolPayInterests(this js.Value, p []js.Value) interface{} {
	blobberID := p[0].String()

	err := sdk.StakePoolPayInterests(blobberID)
	if err != nil {
		fmt.Println("Cannot set wallet info")
	}
	return err
}

// //
// // write pool
// //

// // GetWritePoolInfo for given client, or, if the given clientID is empty,
// // for current client of the sdk.
func ZBOXGetWritePoolInfo(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()
	result, err := sdk.GetWritePoolInfo(clientID)
	if err != nil {
		return err
	}
	return result
}

// // WritePoolLock locks given number of tokes for given duration in read pool.
func WritePoolLock(this js.Value, p []js.Value) interface{} {
	dur, _ := time.ParseDuration(p[0].String()) // time.Duration,
	allocID := p[1].String()
	blobberID := p[2].String()
	tokens, _ := strconv.ParseInt(p[3].String(), 10, 64)
	fee, _ := strconv.ParseInt(p[4].String(), 10, 64)

	err := sdk.WritePoolLock(dur, allocID, blobberID, tokens, fee)
	if err != nil {
		fmt.Println("Cannot set wallet info")
	}
	return err
}

// // WritePoolUnlock unlocks tokens in expired read pool
func WritePoolUnlock(this js.Value, p []js.Value) interface{} {
	poolID := p[0].String()
	fee, _ := strconv.ParseInt(p[1].String(), 10, 64)

	err := sdk.WritePoolUnlock(poolID, fee)
	if err != nil {
		fmt.Println("Cannot set wallet info")
	}
	return err
}

// //
// // challenge pool
// //

// // GetChallengePoolInfo for given allocation.
func ZBOXGetChallengePoolInfo(this js.Value, p []js.Value) interface{} {
	allocID := p[0].String()
	result, err := sdk.GetChallengePoolInfo(allocID)
	if err != nil {
		return err
	}
	return result
}

// //
// // storage SC configurations and blobbers
// //

func ZBOXGetStorageSCConfig(this js.Value, p []js.Value) interface{} {
	result, err := sdk.GetStorageSCConfig()
	if err != nil {
		return err
	}
	return result
}

func ZBOXGetBlobbers(this js.Value, p []js.Value) interface{} {
	result, err := sdk.GetBlobbers()
	if err != nil {
		return err
	}
	return result
}

// // GetBlobber instance.
func ZBOXGetBlobber(this js.Value, p []js.Value) interface{} {
	blobberID := p[0].String()
	result, err := sdk.GetBlobber(blobberID)
	if err != nil {
		return err
	}
	return result
}

// //
// // ---
// //

func ZBOXGetClientEncryptedPublicKey(this js.Value, p []js.Value) interface{} {
	result, err := sdk.GetClientEncryptedPublicKey()
	if err != nil {
		return err
	}
	return result
}

func GetAllocationFromAuthTicket(this js.Value, p []js.Value) interface{} {
	authTicket := p[0].String()
	result, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return err
	}
	return result
}

func ZBOXGetAllocation(this js.Value, p []js.Value) interface{} {
	allocationID := p[0].String()
	result, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return err
	}
	return result
}

func SetNumBlockDownloads(this js.Value, p []js.Value) interface{} {
	num, _ := strconv.ParseInt(p[0].String(), 10, 64)
	sdk.SetNumBlockDownloads(num)
	return nil
}

func ZBOXGetAllocations(this js.Value, p []js.Value) interface{} {
	result, err := sdk.GetAllocations()
	if err != nil {
		return err
	}
	return result
}

func GetAllocationsForClient(this js.Value, p []js.Value) interface{} {
	clientID := p[0].String()
	result, err := sdk.GetAllocationsForClient(clientID)
	if err != nil {
		return err
	}

	return result
}

func CreateAllocation(this js.Value, p []js.Value) interface{} {
	datashards, _ := strconv.Atoi(p[0].String())
	parityshards, _ := strconv.Atoi(p[1].String())
	size, _ := strconv.ParseInt(p[2].String(), 10, 64)
	expiry, _ := strconv.ParseInt(p[3].String(), 10, 64)
	s_read := p[4].String()
	s_write := p[5].String()
	mcct, _ := time.ParseDuration(p[6].String())
	lock, _ := strconv.ParseInt(p[7].String(), 10, 64)

	readPrice := strToPriceRange(s_read)
	writePrice := strToPriceRange(s_write)

	result, err := sdk.CreateAllocation(datashards, parityshards, size, expiry, readPrice, writePrice, mcct, lock)
	if err != nil {
		return err
	}
	return result
}

func CreateAllocationForOwner(this js.Value, p []js.Value) interface{} {
	owner := p[0].String()
	ownerpublickey := p[1].String()
	datashards, _ := strconv.Atoi(p[2].String())
	parityshards, _ := strconv.Atoi(p[3].String())
	size, _ := strconv.ParseInt(p[4].String(), 10, 64)
	expiry, _ := strconv.ParseInt(p[5].String(), 10, 64)
	s_read := p[6].String()
	s_write := p[7].String()
	mcct, _ := time.ParseDuration(p[8].String())
	lock, _ := strconv.ParseInt(p[9].String(), 10, 64)
	preferredBlobbers := ZBOXstrToListSring(p[10].String())

	readPrice := strToPriceRange(s_read)
	writePrice := strToPriceRange(s_write)

	result, err := sdk.CreateAllocationForOwner(owner, ownerpublickey, datashards, parityshards, size, expiry, readPrice, writePrice, mcct, lock, preferredBlobbers)
	if err != nil {
		return err
	}
	return result
}

func UpdateAllocation(this js.Value, p []js.Value) interface{} {
	size, _ := strconv.ParseInt(p[0].String(), 10, 64)
	expiry, _ := strconv.ParseInt(p[1].String(), 10, 64)
	allocationID := p[2].String()
	lock, _ := strconv.ParseInt(p[3].String(), 10, 64)

	result, err := sdk.UpdateAllocation(size, expiry, allocationID, lock)
	if err != nil {
		return err
	}
	return result
}

func FinalizeAllocation(this js.Value, p []js.Value) interface{} {
	allocID := p[0].String()

	result, err := sdk.FinalizeAllocation(allocID)
	if err != nil {
		return err
	}
	return result
}

func CancelAllocation(this js.Value, p []js.Value) interface{} {
	allocID := p[0].String()

	result, err := sdk.CancelAllocation(allocID)
	if err != nil {
		return err
	}
	return result
}

func UpdateBlobberSettings(this js.Value, p []js.Value) interface{} {
	s_blob := p[0].String()
	blob := strToBlob(s_blob)

	result, err := sdk.UpdateBlobberSettings(&blob)
	if err != nil {
		return err
	}
	return result
}

func CommitToFabric(this js.Value, p []js.Value) interface{} {
	metaTxnData := p[0].String()
	fabricConfigJSON := p[1].String()

	result, err := sdk.CommitToFabric(metaTxnData, fabricConfigJSON)
	if err != nil {
		return err
	}
	return result
}

func GetAllocationMinLock(this js.Value, p []js.Value) interface{} {
	datashards, _ := strconv.Atoi(p[0].String())
	parityshards, _ := strconv.Atoi(p[1].String())
	size, _ := strconv.ParseInt(p[2].String(), 10, 64)
	expiry, _ := strconv.ParseInt(p[3].String(), 10, 64)
	s_read := p[4].String()
	s_write := p[5].String()
	mcct, _ := time.ParseDuration(p[6].String())

	readPrice := strToPriceRange(s_read)
	writePrice := strToPriceRange(s_write)

	result, err := sdk.GetAllocationMinLock(datashards, parityshards, size, expiry, readPrice, writePrice, mcct)
	if err != nil {
		return err
	}
	return result
}
