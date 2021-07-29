package blobberClient

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

const BlobberAddr = "localhost:31501"
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

	t.Run("UploadFile", func(t *testing.T) {
		pubKey, privKey, _ := GeneratePubPrivateKey(t)
		allocationTx := randString(32)

		clientId := "exampleOwnerId"

		err := tdController.ClearDatabase()
		if err != nil {
			t.Fatal(err)
		}
		err = tdController.AddUploadTestData(allocationTx, pubKey, clientId)
		if err != nil {
			t.Fatal(err)
		}

		type uploadFormData struct {
			ConnectionID        string `json:"connection_id"`
			Filename            string `json:"filename"`
			Path                string `json:"filepath"`
			Hash                string `json:"content_hash,omitempty"`
			ThumbnailHash       string `json:"thumbnail_content_hash,omitempty"`
			MerkleRoot          string `json:"merkle_root,omitempty"`
			ActualHash          string `json:"actual_hash"`
			ActualSize          int64  `json:"actual_size"`
			ActualThumbnailSize int64  `json:"actual_thumb_size"`
			ActualThumbnailHash string `json:"actual_thumb_hash"`
			MimeType            string `json:"mimetype"`
			CustomMeta          string `json:"custom_meta,omitempty"`
			EncryptedKey        string `json:"encrypted_key,omitempty"`
		}

		formData := uploadFormData{
			ConnectionID:        "exampleconnectionID",
			Filename:            "uploadfile",
			Path:                "somepath",
			ActualHash:          "exampleId:examplePath",
			ActualSize:          102444,
			ActualThumbnailHash: "exampleId:examplePath",
			ActualThumbnailSize: 1024,
			MimeType:            "sometype",
			Hash:                "exampleId:examplePath",
			ThumbnailHash:       "exampleId:examplePath",
			MerkleRoot:          "exmaple:fileMerkleRoot",
		}
		uploadMeta, _ := json.Marshal(formData)

		testCases := []struct {
			name           string
			context        metadata.MD
			clientHeader   string
			input          *blobbergrpc.UploadFileRequest
			expectingError bool
		}{
			{
				name:         "Success",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.UploadFileRequest{
					Path:                "examplePath",
					Allocation:          allocationTx,
					ConnectionId:        "exampleID",
					Method:              "POST",
					UploadMeta:          string(uploadMeta),
					UploadFile:          []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100},
					UploadThumbnailFile: []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100},
				},
				expectingError: false,
			},
			{
				name:         "Unknown file path",
				clientHeader: "exampleOwnerId",
				input: &blobbergrpc.UploadFileRequest{
					Path:                "examplePath",
					Allocation:          allocationTx,
					ConnectionId:        "exampleID",
					Method:              "PUT",
					UpdateMeta:          string(uploadMeta),
					UploadFile:          []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100},
					UploadThumbnailFile: []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100},
				},
				expectingError: false,
			},
		}

		for _, tc := range testCases {
			clientRaw, _ := json.Marshal(client.Client{Wallet: &zcncrypto.Wallet{
				ClientID: "",
				Keys:     []zcncrypto.KeyPair{{PublicKey: pubKey, PrivateKey: privKey}},
			}})

			err = client.PopulateClient(string(clientRaw), signScheme)
			if err != nil {
				t.Fatal(err)
			}

			_, err := UploadFile(BlobberAddr, tc.input)

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
	},
	)
}
