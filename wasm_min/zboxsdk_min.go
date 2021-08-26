package wasm_min

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	result := sdk.InitAuthTicket(authTicket)
	return result
}

func ZBOXGetClientEncryptedPublicKey(this js.Value, p []js.Value) interface{} {
	result, err := sdk.GetClientEncryptedPublicKey()
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
	num, _ := strconv.Atoi(p[0].String())
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

func GetAllocationFromAuthTicket(this js.Value, p []js.Value) interface{} {
	authTicket := p[0].String()
	result, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return err
	}
	return result
}
