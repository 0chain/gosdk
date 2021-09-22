package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/0chain/gosdk/zmagmacore/time"
)

func Test_Billing_CalcAmount(t *testing.T) {
	t.Parallel()

	bill, terms := mockBilling(), mockProviderTerms()

	termsMinCost := mockProviderTerms()
	termsMinCost.MinCost = 1000

	// data usage summary in megabytes
	mbps := float64(bill.DataUsage.UploadBytes+bill.DataUsage.DownloadBytes) / million
	want := int64(mbps * float64(terms.GetPrice()))

	tests := [3]struct {
		name  string
		bill  Billing
		terms ProviderTerms
		want  int64
	}{
		{
			name:  "OK",
			bill:  bill,
			terms: terms,
			want:  want,
		},
		{
			name:  "Zero_Amount_OK",
			bill:  mockBilling(),
			terms: ProviderTerms{},
			want:  0,
		},
		{
			name:  "Min_Cost_Amount_OK",
			bill:  mockBilling(),
			terms: termsMinCost,
			want:  termsMinCost.GetMinCost(),
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.bill.Amount != 0 { // must be zero before first call CalcAmount()
				t.Errorf("Billing.Amount is: %v | want: %v", test.bill.Amount, 0)
			}

			test.bill.CalcAmount(test.terms)
			if test.bill.Amount != test.want { // must be the same value with test.want after called CalcAmount()
				t.Errorf("GetVolume() got: %v | want: %v", test.bill.Amount, test.want)
			}
		})
	}
}

func Test_Billing_Decode(t *testing.T) {
	t.Parallel()

	bill := mockBilling()
	blob, err := json.Marshal(bill)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	billCompleted := mockBilling()
	billCompleted.CalcAmount(mockProviderTerms())
	billCompleted.CompletedAt = time.Now()
	blobCompleted, err := json.Marshal(billCompleted)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [3]struct {
		name  string
		blob  []byte
		want  Billing
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  bill,
			error: false,
		},
		{
			name:  "Completed_OK",
			blob:  blobCompleted,
			want:  billCompleted,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"), // invalid json
			want:  Billing{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := Billing{}
			if err := got.Decode(test.blob); (err != nil) != test.error {
				t.Errorf("Decode() error: %v | want: %v", err, test.error)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Decode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_Billing_Encode(t *testing.T) {
	t.Parallel()

	bill := mockBilling()
	blob, err := json.Marshal(bill)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name string
		bill Billing
		want []byte
	}{
		{
			name: "OK",
			bill: bill,
			want: blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.bill.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_Billing_Validate(t *testing.T) {
	t.Parallel()

	bill, dataUsage := mockBilling(), mockDataUsage()

	duInvalidSessionTime := mockDataUsage()
	duInvalidSessionTime.SessionTime = bill.DataUsage.SessionTime - 1

	duInvalidUploadBytes := mockDataUsage()
	duInvalidUploadBytes.UploadBytes = bill.DataUsage.UploadBytes - 1

	duInvalidDownloadBytes := mockDataUsage()
	duInvalidDownloadBytes.DownloadBytes = bill.DataUsage.DownloadBytes - 1

	tests := [5]struct {
		name  string
		du    *DataUsage
		bill  Billing
		error bool
	}{
		{
			name:  "OK",
			du:    &dataUsage,
			bill:  bill,
			error: false,
		},
		{
			name:  "nil_Data_Usage_ERR",
			du:    nil,
			bill:  bill,
			error: true,
		},
		{
			name:  "Invalid_Session_Time_ERR",
			du:    &duInvalidSessionTime,
			bill:  bill,
			error: true,
		},
		{
			name:  "Invalid_Upload_Bytes_ERR",
			du:    &duInvalidUploadBytes,
			bill:  bill,
			error: true,
		},
		{
			name:  "Invalid_Download_Bytes_ERR",
			du:    &duInvalidDownloadBytes,
			bill:  bill,
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.bill.Validate(test.du); (err != nil) != test.error {
				t.Errorf("validate() error: %v | want: %v", err, test.error)
			}
		})
	}
}
