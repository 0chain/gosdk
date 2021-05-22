package blobberClient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobberHTTP"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/allocation"

	"github.com/0chain/gosdk/core/zcncrypto"

	"github.com/0chain/gosdk/zboxcore/client"

	"google.golang.org/grpc/metadata"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const BlobberAddr = "localhost:7031"
const RetryAttempts = 8
const RetryTimeout = 3
const signScheme = "bls0chain"
const dbName = "blobber_meta"
const dbUser = "blobber_user"
const dbPassword = "blobber"
const dbHost = "localhost"
const dbPort = 5432

func TestBlobberClient_IntegrationTest(t *testing.T) {
	var err error
	args := make(map[string]bool)
	for _, arg := range os.Args {
		args[arg] = true
	}
	if !args["integration"] {
		//t.Skip()
	}

	var conn *grpc.ClientConn
	for i := 0; i < RetryAttempts; i++ {
		log.Println("Connection attempt - " + fmt.Sprint(i+1))
		conn, err = grpc.Dial(BlobberAddr, grpc.WithInsecure())
		if err != nil {
			log.Println(err)
			<-time.After(time.Second * RetryTimeout)
			continue
		}
		break
	}
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%v port=%v user=%v dbname=%v password=%v sslmode=disable",
		dbHost, dbPort,
		dbUser, dbName,
		dbPassword)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	tdController := NewTestDataController(db)

	t.Run("TestGetAllocation", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddGetAllocationTestData()
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name           string
			input          *blobbergrpc.GetAllocationRequest
			expectedTx     string
			expectingError bool
		}{
			{
				name: "Success",
				input: &blobbergrpc.GetAllocationRequest{
					Id: "exampleTransaction",
				},
				expectedTx:     "exampleTransaction",
				expectingError: false,
			},
			{
				name: "UnknownAllocation",
				input: &blobbergrpc.GetAllocationRequest{
					Id: "exampleTransaction1",
				},
				expectedTx:     "",
				expectingError: true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID: "",
				Keys:     []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}

			respRaw, err := GetAllocation(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			var getAllocationResp allocation.Allocation
			json.Unmarshal(respRaw, &getAllocationResp)

			if tc.expectingError {
				t.Fatal("expected error")
			}

			if getAllocationResp.Tx != tc.expectedTx {
				t.Fatal("response with wrong allocation transaction")
			}
		}
	})

	t.Run("TestGetFileMetaData", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddGetFileMetaDataTestData()
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name             string
			context          metadata.MD
			clientHeader     string
			input            *blobbergrpc.GetFileMetaDataRequest
			expectedFileName string
			expectingError   bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.GetFileMetaDataRequest{
					Path:       "examplePath",
					PathHash:   "exampleId:examplePath",
					Allocation: "exampleTransaction",
				},
				expectedFileName: "filename",
				expectingError:   false,
			},
			{
				name:         "Unknown file path",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.GetFileMetaDataRequest{
					Path:       "examplePath",
					PathHash:   "exampleId:examplePath123",
					Allocation: "exampleTransaction",
				},
				expectedFileName: "",
				expectingError:   true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID: tc.clientHeader,
				Keys:     []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}
			respRaw, err := GetFileMetaData(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			var getFileMetaDataResp = make(map[string]interface{})
			json.Unmarshal(respRaw, &getFileMetaDataResp)

			if tc.expectingError {
				t.Fatal("expected error")
			}

			filename, _ := getFileMetaDataResp["name"].(string)

			if filename != tc.expectedFileName {
				t.Fatal("unexpected file name from GetFileMetaData rpc")
			}
		}
	})

	t.Run("TestGetFileStats", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		err = tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddGetFileStatsTestData(allocationTx, pubKey)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name             string
			input            *blobbergrpc.GetFileStatsRequest
			clientHeader     string
			expectedFileName string
			expectingError   bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.GetFileStatsRequest{
					Path:       "examplePath",
					PathHash:   "exampleId:examplePath",
					Allocation: allocationTx,
				},
				expectedFileName: "filename",
				expectingError:   false,
			},
			{
				name:         "Unknown Path",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.GetFileStatsRequest{
					Path:       "examplePath",
					PathHash:   "exampleId:examplePath123",
					Allocation: allocationTx,
				},
				expectedFileName: "",
				expectingError:   true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID: tc.clientHeader,
				Keys:     []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}
			respRaw, err := GetFileStats(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}

			var getFileStatsResp = make(map[string]interface{})
			json.Unmarshal(respRaw, &getFileStatsResp)

			filename, _ := getFileStatsResp["name"].(string)

			if filename != tc.expectedFileName {
				t.Fatal("unexpected file name from GetFileStats rpc")
			}
		}

	})

	t.Run("TestListEntities", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddListEntitiesTestData(allocationTx, pubKey)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name           string
			clientHeader   string
			input          *blobbergrpc.ListEntitiesRequest
			expectedPath   string
			expectingError bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.ListEntitiesRequest{
					Path:       "examplePath",
					PathHash:   "exampleId:examplePath",
					AuthToken:  "",
					Allocation: allocationTx,
				},
				expectedPath:   "examplePath",
				expectingError: false,
			},
			{
				name:         "bad path",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.ListEntitiesRequest{
					Path:       "examplePath",
					PathHash:   "exampleId:examplePath123",
					AuthToken:  "",
					Allocation: allocationTx,
				},
				expectedPath:   "",
				expectingError: true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID: tc.clientHeader,
				Keys:     []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}
			respRaw, err := ListEntities(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}

			var listEntitiesResp blobberHTTP.ListResult
			json.Unmarshal(respRaw, &listEntitiesResp)

			if path, _ := listEntitiesResp.Meta["path"].(string); path != tc.expectedPath {
				t.Fatal("unexpected path from ListEntities rpc")
			}
		}

	})

	t.Run("TestGetReferencePath", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddGetReferencePathTestData(allocationTx, pubKey)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name           string
			clientHeader   string
			input          *blobbergrpc.GetReferencePathRequest
			expectedPath   string
			expectingError bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.GetReferencePathRequest{
					Paths:      "",
					Path:       "/",
					Allocation: allocationTx,
				},
				expectedPath:   "/",
				expectingError: false,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID: tc.clientHeader,
				Keys:     []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}

			respRaw, err := GetReferencePath(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}

			var getReferencePathResp *blobberHTTP.ReferencePathResult
			json.Unmarshal(respRaw, &getReferencePathResp)

			if path, _ := getReferencePathResp.ReferencePath.Meta["path"].(string); path != tc.expectedPath {
				t.Fatal("unexpected path from GetReferencePath rpc")
			}
		}
	})

	t.Run("TestGetObjectTree", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddGetObjectTreeTestData(allocationTx, pubKey)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name             string
			clientHeader     string
			input            *blobbergrpc.GetObjectTreeRequest
			expectedFileName string
			expectingError   bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.GetObjectTreeRequest{
					Path:       "/",
					Allocation: allocationTx,
				},
				expectedFileName: "root",
				expectingError:   false,
			},
			{
				name:         "bad path",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.GetObjectTreeRequest{
					Path:       "/2",
					Allocation: "",
				},
				expectedFileName: "root",
				expectingError:   true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID: tc.clientHeader,
				Keys:     []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}
			respRaw, err := GetObjectTree(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}

			var getObjectTreeResp blobberHTTP.ReferencePathResult
			json.Unmarshal(respRaw, &getObjectTreeResp)

			if filename, _ := getObjectTreeResp.ReferencePath.Meta["name"].(string); filename != tc.expectedFileName {
				t.Fatal("unexpected root name from GetObject")
			}
		}

	})

}
