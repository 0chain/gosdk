package magmasc

import (
	"encoding/hex"
	"time"

	magma "github.com/magma/augmented-networks/accounting/protos"
	"golang.org/x/crypto/sha3"

	"github.com/0chain/gosdk/zmagmacore/crypto"
	"github.com/0chain/gosdk/zmagmacore/magmasc/pb"
	ts "github.com/0chain/gosdk/zmagmacore/time"
)

func mockSession() *Session {
	now := time.Now().Format(time.RFC3339Nano)
	billing := mockBilling()

	return &Session{
		SessionID:   billing.DataMarker.DataUsage.SessionId,
		AccessPoint: &AccessPoint{ID: "id:access:point:" + now},
		Billing:     billing,
		Consumer:    mockConsumer(),
		Provider:    mockProvider(),
	}
}

func mockBilling() Billing {
	return Billing{
		DataMarker: mockDataMarker(),
	}
}

func mockConsumer() *Consumer {
	now := time.Now().Format(time.RFC3339Nano)
	return &Consumer{
		ID:    "id:consumer:" + now,
		ExtID: "id:consumer:external:" + now,
		Host:  "localhost:8010",
	}
}

func mockDataUsage() *pb.DataUsage {
	now := time.Now().Format(time.RFC3339Nano)
	return &pb.DataUsage{
		DownloadBytes: 3 * million,
		UploadBytes:   2 * million,
		SessionId:     "id:session:" + now,
		SessionTime:   1 * 60, // 1 minute
	}
}

func mockProvider() *Provider {
	now := time.Now().Format(time.RFC3339Nano)
	return &Provider{
		&pb.Provider{
			ID:    "id:provider:" + now,
			ExtID: "id:provider:external:" + now,
			Host:  "localhost:8020",
		},
	}
}

func mockTerms() Terms {
	return Terms{
		Price:           0.1,
		PriceAutoUpdate: 0.001,
		MinCost:         0.5,
		Volume:          0,
		QoS:             mockQoS(),
		QoSAutoUpdate: &QoSAutoUpdate{
			DownloadMbps: 0.001,
			UploadMbps:   0.001,
		},
		ProlongDuration: 1 * 60 * 60,              // 1 hour
		ExpiredAt:       ts.Now() + (1 * 60 * 60), // 1 hour from now
	}
}

func mockTokenPool() *TokenPool {
	now := time.Now().Format(time.RFC3339Nano)
	return &TokenPool{
		ID:       "id:session:" + now,
		Balance:  1000,
		HolderID: "id:holder:" + now,
		PayerID:  "id:payer:" + now,
		PayeeID:  "id:payee:" + now,
		Transfers: []TokenPoolTransfer{
			mockTokenPoolTransfer(),
			mockTokenPoolTransfer(),
			mockTokenPoolTransfer(),
		},
	}
}

func mockTokenPoolTransfer() TokenPoolTransfer {
	now := time.Now()
	bin, _ := time.Now().MarshalBinary()
	hash := sha3.Sum256(bin)
	fix := now.Format(time.RFC3339Nano)

	return TokenPoolTransfer{
		TxnHash:    hex.EncodeToString(hash[:]),
		FromPool:   "id:from:pool:" + fix,
		ToPool:     "id:to:pool:" + fix,
		Value:      1111,
		FromClient: "id:from:client:" + fix,
		ToClient:   "id:to:client:" + fix,
	}
}

func mockQoS() *magma.QoS {
	return &magma.QoS{
		DownloadMbps: 5.4321,
		UploadMbps:   1.2345,
	}
}

func mockDataMarker() *DataMarker {
	pbKey := "87600ad70ce2e902e6b5d67762b6d623a3d3a3caedd300529dab947b72e2c813874a2635c1c7bc0226d16031c2575c1f6c8094cea8ebfe213a9c8fe20deb4695"
	// private key for public key above
	// pr := "42b649226c17b6b6f03d5e3f5c63a311ba0d520ad18188a1a0d79324885a051a"
	return &DataMarker{
		DataMarker: &pb.DataMarker{
			UserId:    crypto.Hash(pbKey),
			DataUsage: mockDataUsage(),
			Qos: &pb.QoS{
				DownloadMbps: 5.4321,
				UploadMbps:   1.2345,
				Latency:      6.789,
			},
			PublicKey: pbKey,
			SigScheme: "bls0chain",
			Signature: "da40ea8816a242f205d4e3b6cfad3dcf43eaa617bf755ed69788ecd80ffec98b",
		},
	}
}

func mockUser() *User {
	now := time.Now().Format(time.RFC3339Nano)
	return &User{
		User: &pb.User{
			ID:         "id:user:" + now,
			ConsumerID: "id:consumer:" + now,
		},
	}
}
