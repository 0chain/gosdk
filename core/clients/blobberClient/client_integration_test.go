package blobberClient

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/readmarker"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobberhttp"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/allocation"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/reference"
	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/writemarker"
	"github.com/0chain/blobber/code/go/0chain.net/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const BlobberAddr = "https://localhost:31501"
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
		t.Skip()
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

			var listEntitiesResp blobberhttp.ListResult
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

			var getReferencePathResp *blobberhttp.ReferencePathResult
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

			var getObjectTreeResp blobberhttp.ReferencePathResult
			json.Unmarshal(respRaw, &getObjectTreeResp)

			if filename, _ := getObjectTreeResp.ReferencePath.Meta["name"].(string); filename != tc.expectedFileName {
				t.Fatal("unexpected root name from GetObject")
			}
		}

	})

	t.Run("TestCommit", func(t *testing.T) {
		allocationTx := randString(32)

		pubKey, privKey, ss := GeneratePubPrivateKey(t)
		pubKeyBytes, _ := hex.DecodeString(pubKey)
		clientId := encryption.Hash(pubKeyBytes)
		now := common.Timestamp(time.Now().UnixNano())

		blobberPubKey := "de52c0a51872d5d2ec04dbc15a6f0696cba22657b80520e1d070e72de64c9b04e19ce3223cae3c743a20184158457582ffe9c369ca9218c04bfe83a26a62d88d"
		blobberPubKeyBytes, _ := hex.DecodeString(blobberPubKey)

		fr := reference.Ref{
			AllocationID:   "exampleId",
			Type:           "f",
			Name:           "new_name",
			Path:           "/new_name",
			ContentHash:    "contentHash",
			MerkleRoot:     "merkleRoot",
			ActualFileHash: "actualFileHash",
		}

		rootRefHash := encryption.Hash(encryption.Hash(fr.GetFileHashData()))

		wm := writemarker.WriteMarker{
			AllocationRoot:         encryption.Hash(rootRefHash + ":" + strconv.FormatInt(int64(now), 10)),
			PreviousAllocationRoot: "/",
			AllocationID:           "exampleId",
			Size:                   1337,
			BlobberID:              encryption.Hash(blobberPubKeyBytes),
			Timestamp:              now,
			ClientID:               clientId,
		}

		wmSig, err := ss.Sign(encryption.Hash(wm.GetHashData()))
		if err != nil {
			t.Fatal(err)
		}

		wm.Signature = wmSig

		wmRaw, err := json.Marshal(wm)
		if err != nil {
			t.Fatal(err)
		}

		err = tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddCommitTestData(allocationTx, pubKey, clientId, wmSig, now)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name               string
			input              *blobbergrpc.CommitRequest
			expectedAllocation string
			expectingError     bool
		}{
			{
				name: "Success",
				input: &blobbergrpc.CommitRequest{
					Allocation:   allocationTx,
					ConnectionId: "connection_id",
					WriteMarker:  string(wmRaw),
				},
				expectedAllocation: "exampleId",
				expectingError:     false,
			},
			{
				name: "invalid write_marker",
				input: &blobbergrpc.CommitRequest{
					Allocation:   allocationTx,
					ConnectionId: "invalid",
					WriteMarker:  "invalid",
				},
				expectedAllocation: "",
				expectingError:     true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID:  clientId,
				ClientKey: pubKey,
				Keys:      []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}
			respRaw, err := Commit(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}

			var commiteResp blobberhttp.CommitResult
			json.Unmarshal(respRaw, &commiteResp)

			if commiteResp.WriteMarker.AllocationID != tc.expectedAllocation {
				t.Fatal("unexpected allocationId")
			}
		}
	})

	t.Run("CommitMetaTxn", func(t *testing.T) {
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
			input          *blobbergrpc.CommitMetaTxnRequest
			expectedPath   string
			expectingError bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.CommitMetaTxnRequest{
					Path:       "examplePath",
					PathHash:   "exampleId:examplePath",
					AuthToken:  "",
					Allocation: allocationTx,
					TxnId:      "8",
				},
				expectedPath:   "examplePath",
				expectingError: false,
			},
			{
				name:         "bad path",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.CommitMetaTxnRequest{
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
			_, err = CommitMetaTxn(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}
		}
	})

	t.Run("TestCollaborator", func(t *testing.T) {
		allocationTx := randString(32)

		pubKey, privKey, ss := GeneratePubPrivateKey(t)
		pubKeyBytes, _ := hex.DecodeString(pubKey)
		clientId := encryption.Hash(pubKeyBytes)
		now := common.Timestamp(time.Now().UnixNano())

		blobberPubKey := "de52c0a51872d5d2ec04dbc15a6f0696cba22657b80520e1d070e72de64c9b04e19ce3223cae3c743a20184158457582ffe9c369ca9218c04bfe83a26a62d88d"
		blobberPubKeyBytes, _ := hex.DecodeString(blobberPubKey)

		fr := reference.Ref{
			AllocationID:   "exampleId",
			Type:           "f",
			Name:           "new_name",
			Path:           "/new_name",
			ContentHash:    "contentHash",
			MerkleRoot:     "merkleRoot",
			ActualFileHash: "actualFileHash",
		}

		rootRefHash := encryption.Hash(encryption.Hash(fr.GetFileHashData()))

		wm := writemarker.WriteMarker{
			AllocationRoot:         encryption.Hash(rootRefHash + ":" + strconv.FormatInt(int64(now), 10)),
			PreviousAllocationRoot: "/",
			AllocationID:           "exampleId",
			Size:                   1337,
			BlobberID:              encryption.Hash(blobberPubKeyBytes),
			Timestamp:              now,
			ClientID:               clientId,
		}

		wmSig, err := ss.Sign(encryption.Hash(wm.GetHashData()))
		if err != nil {
			t.Fatal(err)
		}

		wm.Signature = wmSig

		err = tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddCommitTestData(allocationTx, pubKey, clientId, wmSig, now)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name               string
			input              *blobbergrpc.CollaboratorRequest
			expectedAllocation string
			expectingError     bool
		}{
			{
				name: "Success",
				input: &blobbergrpc.CollaboratorRequest{
					Allocation: allocationTx,
					CollabId:   "10",
					Method:     http.MethodPost,
					Path:       "/some_file",
					PathHash:   "exampleId:examplePath",
				},
				expectedAllocation: "exampleId",
				expectingError:     false,
			},
			{
				name: "invalid write_marker",
				input: &blobbergrpc.CollaboratorRequest{
					Path:       "/some_file",
					PathHash:   "exampleId:examplePath",
					Allocation: allocationTx,
					Method:     http.MethodPost,
				},
				expectedAllocation: "",
				expectingError:     true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID:  clientId,
				ClientKey: pubKey,
				Keys:      []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Collaborator(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}
		}
	})

	t.Run("TestCalculateHash", func(t *testing.T) {
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
			input          *blobbergrpc.CalculateHashRequest
			expectedPath   string
			expectingError bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.CalculateHashRequest{
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

			_, err = CalculateHash(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}

			if tc.expectingError {
				t.Fatal("expected error")
			}
		}
	})

	t.Run("TestUpdateObjectAttributes", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		pubKeyBytes, _ := hex.DecodeString(pubKey)
		clientId := encryption.Hash(pubKeyBytes)

		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddAttributesTestData(allocationTx, pubKey, clientId)
		if err != nil {
			t.Fatal(err)
		}

		attr := &reference.Attributes{WhoPaysForReads: common.WhoPays3rdParty}
		attrBytes, err := json.Marshal(attr)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name           string
			clientHeader   string
			input          *blobbergrpc.UpdateObjectAttributesRequest
			expectedPath   string
			expectingError bool
		}{
			{
				name: "Success",
				input: &blobbergrpc.UpdateObjectAttributesRequest{
					Path:         "/",
					PathHash:     "exampleId:examplePath",
					Allocation:   allocationTx,
					ConnectionId: "connection_id",
					Attributes:   string(attrBytes),
				},
				expectedPath:   "/",
				expectingError: false,
			},
			{
				name: "Failed",
				input: &blobbergrpc.UpdateObjectAttributesRequest{
					Path:         "",
					PathHash:     "",
					Allocation:   "",
					ConnectionId: "",
					Attributes:   "",
				},
				expectedPath:   "/",
				expectingError: true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID:  clientId,
				ClientKey: pubKey,
				Keys:      []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}

			_, err = UpdateObjectAttributes(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}
			if tc.expectingError {
				t.Fatal("expected error")
			}
		}
	})

	t.Run("TestCopyObject", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		pubKeyBytes, _ := hex.DecodeString(pubKey)
		clientId := encryption.Hash(pubKeyBytes)

		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddCopyObjectData(allocationTx, pubKey, clientId)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name            string
			clientHeader    string
			input           *blobbergrpc.CopyObjectRequest
			expectedMessage string
			expectingError  bool
		}{
			{
				name: "Success",
				input: &blobbergrpc.CopyObjectRequest{
					Path:         "/some_file",
					PathHash:     "exampleId:examplePath",
					Allocation:   allocationTx,
					ConnectionId: "connection_id",
					Dest:         "/copy",
				},
				expectedMessage: "some_file",
				expectingError:  false,
			},
			{
				name: "Failed",
				input: &blobbergrpc.CopyObjectRequest{
					Path:         "",
					PathHash:     "",
					Allocation:   "",
					ConnectionId: "",
					Dest:         "",
				},
				expectedMessage: "",
				expectingError:  true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID:  clientId,
				ClientKey: pubKey,
				Keys:      []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}

			response, err := CopyObject(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}
			if tc.expectingError {
				t.Fatal("expected error")
			}

			copyResponseResult := &blobberhttp.UploadResult{}
			err = json.Unmarshal(response, copyResponseResult)
			if err != nil {
				t.Fatalf("failed to unmarshal %+v", err)
			}
			if copyResponseResult.Filename != tc.expectedMessage {
				t.Fatalf("Invalid filename, expected %s got %s", copyResponseResult.Filename, tc.expectedMessage)
			}
		}
	})

	t.Run("TestRenameObject", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		pubKeyBytes, _ := hex.DecodeString(pubKey)
		clientId := encryption.Hash(pubKeyBytes)

		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddRenameTestData(allocationTx, pubKey, clientId)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name           string
			clientHeader   string
			input          *blobbergrpc.RenameObjectRequest
			expectedPath   string
			expectingError bool
		}{
			{
				name: "Success",
				input: &blobbergrpc.RenameObjectRequest{
					Path:         "/",
					PathHash:     "exampleId:examplePath",
					Allocation:   allocationTx,
					ConnectionId: "connection_id",
					NewName:      "somethingNew",
				},
				expectedPath:   "/",
				expectingError: false,
			},
			{
				name: "Failed",
				input: &blobbergrpc.RenameObjectRequest{
					Path:         "",
					PathHash:     "",
					Allocation:   "",
					ConnectionId: "",
					NewName:      "",
				},
				expectedPath:   "/",
				expectingError: true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID:  clientId,
				ClientKey: pubKey,
				Keys:      []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}

			_, err = RenameObject(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}
			if tc.expectingError {
				t.Fatal("expected error")
			}
		}
	})

	t.Run("TestDownloadObject", func(t *testing.T) {
		pubKey, privKey, signSch := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		root, _ := os.Getwd()
		path := strings.Split(root, `core`)
		fmt.Println("PATH IS ", path[0], path[1])
		//_ = os.Di("../../../")


		err := os.MkdirAll(`/blobber/files/files/exa/mpl/eId/objects/tmp/Mon/Wen`, 0700)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := os.RemoveAll(`/blobber/files/files/exa/mpl/eId/objects/tmp/Mon`)
			if err != nil {
				t.Fatal(err)
			}
		}()

		f, err := os.Create(`/blobber/files/files/exa/mpl/eId/objects/tmp/Mon/Wen/MyFile`)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		//file, err := os.Open(root + "/helper_integration_test.go")
		//if err != nil {
		//	t.Fatal(err)
		//}
		//defer file.Close()
		//
		//_, err = io.Copy(f, file)
		//if err != nil {
		//	t.Fatal(err)
		//}

		pubKeyBytes, _ := hex.DecodeString(pubKey)
		clientId := encryption.Hash(pubKeyBytes)

		err = tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		allocationId := `exampleId`
		now := common.Timestamp(time.Now().Unix())

		blobberPubKey := "de52c0a51872d5d2ec04dbc15a6f0696cba22657b80520e1d070e72de64c9b04e19ce3223cae3c743a20184158457582ffe9c369ca9218c04bfe83a26a62d88d"
		blobberPubKeyBytes, _ := hex.DecodeString(blobberPubKey)

		rm := readmarker.ReadMarker{
			BlobberID:       encryption.Hash(blobberPubKeyBytes),
			AllocationID:    allocationId,
			ClientPublicKey: pubKey,
			ClientID:        clientId,
			OwnerID:         clientId,
			Timestamp:       now,
			//ReadCounter:     1337,
		}

		rmSig, err := signSch.Sign(encryption.Hash(rm.GetHashData()))
		if err != nil {
			t.Fatal(err)
		}
		rm.Signature = rmSig

		rmString, err := json.Marshal(rm)
		if err != nil {
			t.Fatal(err)
		}

		err = tdController.AddDownloadTestData(allocationTx, pubKey, clientId, rmSig, now)
		if err != nil {
			t.Fatal(err)
		}

		testCases := []struct {
			name           string
			clientHeader   string
			input          *blobbergrpc.DownloadFileRequest
			expectedPath   string
			expectingError bool
		}{
			{
				name: "Success",
				input: &blobbergrpc.DownloadFileRequest{
					Allocation: allocationTx,
					Path:       "/some_file",
					PathHash:   "exampleId:examplePath",
					ReadMarker: string(rmString),
					BlockNum:   "1",
				},
				expectedPath:   "some_file",
				expectingError: false,
			},
			{
				name: "Failed",
				input: &blobbergrpc.DownloadFileRequest{
					Allocation: "",
					Path:       "",
					PathHash:   "",
					RxPay:      "",
					BlockNum:   "",
					NumBlocks:  "",
					ReadMarker: "",
					AuthToken:  "",
					Content:    "",
				},
				expectedPath:   "",
				expectingError: true,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID:  clientId,
				ClientKey: pubKey,
				Keys:      []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err := client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}

			_, err = DownloadObject(BlobberAddr, tc.input)
			if err != nil {
				if !tc.expectingError {
					t.Fatal(err)
				}
				continue
			}
			if tc.expectingError {
				t.Fatal("expected error")
			}
		}
	})
}
