package util_test

import (
	"github.com/0chain/gosdk/util"
	// "0chain.net/zcn"
	// "fmt"
	// "io/ioutil"
	"testing"
)

const NUMCONNID = 16

// Create NUMCONNID and make sure they are unique
func TestNewConnectionId(t *testing.T) {
	t.Log("Create New connection ", NUMCONNID, "ids (expected all to be unique)")
	connid := make([]int64, NUMCONNID)
	for i := 0; i < NUMCONNID; i++ {
		connid[i] = util.NewConnectionId()
	}
	for i := 0; i < NUMCONNID; i++ {
		for j := i + 1; j < NUMCONNID; j++ {
			if connid[i] == connid[j] {
				t.Fail()
			}
		}
	}
}

var clientStr string = `{
	"id" : "8cd930c50b8e06d9ba2ab6a86ca9e3c6d073974d6976312f36a766a7443efd55",
	"public_key" : "78d4cd4d6edbfb3f0a1c7dec479d5b295b9abc2c1d2d332e67a115a62a3c1fd0",
	"private_key" : "c8e05e590c3beddf0c2a239d04a92c20323e660d92e9d2a096e46577f4595b14"
}`

func TestGetBlobbers(t *testing.T) {
	blobStr :=
		`[
		  {
			"id": "blobber1",
			"url": "http://blobber1.com/"
		  },
		  {
			"id": "blobber2",
			"url": "http://blobber2.com/"
		  }
	 ]`
	blobbers, err := util.GetBlobbers(blobStr)
	if err != nil || len(blobbers) != 2 {
		t.Fail()
	}
	if blobbers[0].Id != "blobber1" || blobbers[0].UrlRoot != "http://blobber1.com/" {
		t.Fail()
	}
	if blobbers[1].Id != "blobber2" || blobbers[1].UrlRoot != "http://blobber2.com/" {
		t.Fail()
	}
}

func TestGetClientConfig(t *testing.T) {

	client, err := util.GetClientConfig(clientStr)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if client.Id != "8cd930c50b8e06d9ba2ab6a86ca9e3c6d073974d6976312f36a766a7443efd55" || client.PublicKey != "78d4cd4d6edbfb3f0a1c7dec479d5b295b9abc2c1d2d332e67a115a62a3c1fd0" || client.PrivateKey != "c8e05e590c3beddf0c2a239d04a92c20323e660d92e9d2a096e46577f4595b14" {
		t.Fatalf("Client param doesn't match %v", client)
	}
}


// func TestWorker(t *testing.T) {
// 	util.CreateWorker()
// 	for i := 0; i < 10; i++ {
// 		util.EnqueueJob(i)
// 	}
// 	util.WaitJobs()
// }
