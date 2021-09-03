package magmasc

import (
	"encoding/hex"
	"time"

	magma "github.com/magma/augmented-networks/accounting/protos"
	"golang.org/x/crypto/sha3"

	ts "github.com/0chain/gosdk/zmagmacore/time"
)

func mockAcknowledgment() *Acknowledgment {
	now := time.Now().Format(time.RFC3339Nano)
	billing := mockBilling()

	return &Acknowledgment{
		SessionID:     billing.DataUsage.SessionID,
		AccessPointID: "id:access:point:" + now,
		Billing:       billing,
		Consumer:      mockConsumer(),
		Provider:      mockProvider(),
	}
}

func mockBilling() Billing {
	return Billing{
		DataUsage: mockDataUsage(),
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

func mockDataUsage() DataUsage {
	now := time.Now().Format(time.RFC3339Nano)
	return DataUsage{
		DownloadBytes: 3 * million,
		UploadBytes:   2 * million,
		SessionID:     "id:session:" + now,
		SessionTime:   1 * 60, // 1 minute
	}
}

func mockProvider() *Provider {
	now := time.Now().Format(time.RFC3339Nano)
	return &Provider{
		ID:    "id:provider:" + now,
		ExtID: "id:provider:external:" + now,
		Host:  "localhost:8020",
		MinStake: billion,
	}
}

func mockProviderTerms() ProviderTerms {
	return ProviderTerms{
		AccessPointID:   "id:access:point" + time.Now().Format(time.RFC3339Nano),
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
