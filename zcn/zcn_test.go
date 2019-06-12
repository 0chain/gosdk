package zcn

import (
	// "bytes"
	// "fmt"
	// "io"
	// "math"
	// "os"
	// "path/filepath"
	"testing"
)

var blobStr string = `[
	{
	"Id": "ccde4b06d02e24113889164d94a9692284f60c8701f10e600bde58889d335055",
	"UrlRoot": "http://localhost:5051",
	"PrevRootHash": "123456"
	},
	{
	"Id": "f45305566d865b06ff045e877d6daa76425ee04c84053d35aa462c4962351f45",
	"UrlRoot": "http://localhost:5052",
	"PrevRootHash": "7891011"
	},
	{
	"Id": "73c154fc8347296880da8eaa9a8ad428d43196b438ceb28e64de1e19724e7e20",
	"UrlRoot": "http://localhost:5053",
	"PrevRootHash": "7891011"
	}
]`

var clientStr string = `{
	"Id" : "8cd930c50b8e06d9ba2ab6a86ca9e3c6d073974d6976312f36a766a7443efd55",
	"PublicKey" : "78d4cd4d6edbfb3f0a1c7dec479d5b295b9abc2c1d2d332e67a115a62a3c1fd0",
	"PrivateKey" : "c8e05e590c3beddf0c2a239d04a92c20323e660d92e9d2a096e46577f4595b1478d4cd4d6edbfb3f0a1c7dec479d5b295b9abc2c1d2d332e67a115a62a3c1fd0"
}`

const FILE_TO_UPLOAD = "/tmp/googlechrome.dmg"

const NUMDATASHARDS = 2
const NUMPARITYSHARDS = 1

var TOTALSHARDS int = NUMDATASHARDS + NUMPARITYSHARDS

const CONTENTSIZE = 100

func TestErasureCode(t *testing.T) {
	t.Log("Create Encoder, encode and decode")
	enc, err := newEncoder(NUMDATASHARDS, NUMPARITYSHARDS)
	if err != nil || enc == nil {
		t.Fail()
	}
	data := make([]byte, CONTENTSIZE)
	for i := 0; i < CONTENTSIZE; i++ {
		data[i] = byte(i)
	}
	var shards [][]byte
	shards, err = enc.encode(data)
	if err != nil || len(shards) != TOTALSHARDS {
		t.Fail()
	}
	// Miss any 2 of shards and try to recover
	testShard := make([][]byte, TOTALSHARDS)
	for i := 1; i < TOTALSHARDS; i++ {
		testShard[i] = make([]byte, len(shards[i]))
		copy(testShard[i], shards[i])
	}
	var result []byte
	result, err = enc.decode(testShard, len(shards[0]))
	if err != nil || len(result) < CONTENTSIZE {
		t.Fail()
	}
}

// func TestBlobUpload(t *testing.T) {
// 	alloc, err := CreateInstance("b1225f315f0bfc730bba56d5905f186269828d25cc3b4fd11eaf168777709c98")
// 	err   = alloc.SetConfig(clientStr, blobStr, NUMDATASHARDS, NUMPARITYSHARDS)
// 	if err != nil {
// 		t.Fatal()
// 	}

// 	_, err = os.Stat(FILE_TO_UPLOAD)
// 	if err != nil {
// 		fmt.Println(err)
// 		t.Fatal()
// 	}
// 	var inputFile *os.File
// 	inputFile, err = os.Open(FILE_TO_UPLOAD)
// 	if err != nil {
// 		t.Fatal()
// 	}
// 	defer inputFile.Close()
// 	var inputStat os.FileInfo
// 	inputStat, err = inputFile.Stat()
// 	if err != nil {
// 		t.Fatal()
// 	}

// 	fileStr := fmt.Sprintf("{\"Name\": \"%s\",\"Path\": \"/\",\"Size\": %d,\"Type\": \"%s\"}",
// 				filepath.Base(FILE_TO_UPLOAD), inputStat.Size(), filepath.Ext(FILE_TO_UPLOAD))

// 	err = alloc.SetFileConfig(fileStr)
// 	if err != nil {
// 		t.Fail()
// 	}

// 	size := inputStat.Size()
// 	// Calculate number of bytes per shard.
// 	perShard := (size + int64(NUMDATASHARDS) - 1) / int64(NUMDATASHARDS)
// 	// Pad data to Shards*perShard.
// 	padding := make([]byte, (int64(NUMDATASHARDS)*perShard)-size)
// 	dataReader := io.MultiReader(inputFile, bytes.NewBuffer(padding))
// 	chunksPerShard := (perShard + int64(CHUNK_SIZE) - 1) / CHUNK_SIZE
// 	fmt.Println("Size:", size, "perShard:",perShard, "chunks/shard:", chunksPerShard)

// 	for ctr := int64(0); ctr < chunksPerShard; ctr++ {
// 		remaining := int64(math.Min(float64(perShard-(ctr*CHUNK_SIZE)), CHUNK_SIZE))
// 		// fmt.Println("[", ctr,"] Reading", remaining*int64(NUMDATASHARDS), "bytes...")
// 		b1 := make([]byte, remaining*int64(NUMDATASHARDS))
// 		_, err = dataReader.Read(b1)
// 		if err != nil {
// 			fmt.Println("Read failed", err)
// 			t.Fatal()
// 		}
// 		// fmt.Println("Read", len(b1))
// 		err = alloc.Push(b1)
// 		if err != nil {
// 			t.Fatalf("%s", err.Error())
// 		}
// 	}
// 	err = alloc.CompletePush()
// 	if err != nil {
// 		fmt.Println("Upload failed",err)
// 		t.Fatal()
// 	}
// }
