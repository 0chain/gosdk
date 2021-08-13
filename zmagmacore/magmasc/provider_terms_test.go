package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/0chain/gosdk/zmagmacore/time"
)

func Test_ProviderTerms_Decode(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()
	blob, err := json.Marshal(terms)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	termsInvalid := mockProviderTerms()
	termsInvalid.QoS.UploadMbps = -0.1
	uBlobInvalid, err := json.Marshal(termsInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	termsInvalid = mockProviderTerms()
	termsInvalid.QoS.DownloadMbps = -0.1
	dBlobInvalid, err := json.Marshal(termsInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [4]struct {
		name  string
		blob  []byte
		want  *ProviderTerms
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  &terms,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"), // invalid json
			want:  NewProviderTerms(),
			error: true,
		},
		{
			name:  "QoS_Upload_Mbps_Invalid_ERR",
			blob:  uBlobInvalid,
			want:  NewProviderTerms(),
			error: true,
		},
		{
			name:  "QoS_Download_Mbps_Invalid_ERR",
			blob:  dBlobInvalid,
			want:  NewProviderTerms(),
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := NewProviderTerms()
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

func Test_ProviderTerms_Decrease(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()
	// the volume of terms must to be zeroed
	terms.Volume = 0
	// prolong terms expire
	terms.ExpiredAt += time.Timestamp(terms.ProlongDuration)
	if terms.PriceAutoUpdate != 0 && terms.Price > terms.PriceAutoUpdate {
		terms.Price -= terms.PriceAutoUpdate // down the price
	}
	if terms.QoSAutoUpdate.UploadMbps != 0 {
		terms.QoS.UploadMbps += terms.QoSAutoUpdate.UploadMbps // up the qos of upload mbps
	}
	if terms.QoSAutoUpdate.DownloadMbps != 0 {
		terms.QoS.DownloadMbps += terms.QoSAutoUpdate.DownloadMbps // up the qos of download mbps
	}

	tests := [1]struct {
		name  string
		terms ProviderTerms
		want  ProviderTerms
	}{
		{
			name:  "OK",
			terms: mockProviderTerms(),
			want:  terms,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			test.terms.Decrease()
			if !reflect.DeepEqual(test.terms, test.want) {
				t.Errorf("Decrease() got: %#v | want: %#v", test.terms, test.want)
			}
		})
	}
}

func Test_ProviderTerms_Encode(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()
	blob, err := json.Marshal(terms)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name  string
		terms ProviderTerms
		want  []byte
	}{
		{
			name:  "OK",
			terms: terms,
			want:  blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.terms.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_ProviderTerms_GetAmount(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()

	termsZeroPrice := mockProviderTerms()
	termsZeroPrice.Price = 0

	termsMinCost := mockProviderTerms()
	termsMinCost.Price = 1e-9

	tests := [3]struct {
		name  string
		terms ProviderTerms
		want  int64
	}{
		{
			name:  "OK",
			terms: terms,
			want:  terms.GetPrice() * terms.GetVolume(),
		},
		{
			name:  "Zero_OK",
			terms: termsZeroPrice,
			want:  0,
		},
		{
			name:  "Min_Cost_OK",
			terms: termsMinCost,
			want:  termsMinCost.GetMinCost(),
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.terms.GetAmount(); got != test.want {
				t.Errorf("GetAmount() got: %v | want: %v", got, test.want)
			}
		})
	}
}

func Test_ProviderTerms_GetMinCost(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()
	minCost := int64(terms.MinCost * billion)

	termsZeroMinCost := mockProviderTerms()
	termsZeroMinCost.MinCost = 0

	tests := [2]struct {
		name  string
		terms ProviderTerms
		want  int64
	}{
		{
			name:  "OK",
			terms: terms,
			want:  minCost,
		},
		{
			name:  "Zero_Min_Cost_OK",
			terms: termsZeroMinCost,
			want:  0,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.terms.GetMinCost(); got != test.want {
				t.Errorf("GetMinCost() got: %v | want: %v", got, test.want)
			}
		})
	}
}

func Test_ProviderTerms_GetPrice(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()
	price := int64(terms.Price * billion)

	termsZeroPrice := mockProviderTerms()
	termsZeroPrice.Price = 0

	tests := [2]struct {
		name  string
		terms ProviderTerms
		want  int64
	}{
		{
			name:  "OK",
			terms: terms,
			want:  price,
		},
		{
			name:  "Zero_Price_OK",
			terms: termsZeroPrice,
			want:  0,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.terms.GetPrice(); got != test.want {
				t.Errorf("GetPrice() got: %v | want: %v", got, test.want)
			}
		})
	}
}

func Test_ProviderTerms_GetVolume(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()

	mbps := (terms.QoS.UploadMbps + terms.QoS.DownloadMbps) / octet // mega bytes per second
	duration := float32(terms.ExpiredAt - time.Now())               // duration in seconds
	// rounded of bytes per second multiplied by duration in seconds
	volume := int64(mbps * duration)

	tests := [1]struct {
		name  string
		terms ProviderTerms
		want  int64
	}{
		{
			name:  "OK",
			terms: terms,
			want:  volume,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.terms.Volume != 0 { // must be zero before first call GetVolume()
				t.Errorf("ProviderTerms.Volume is: %v | want: %v", test.terms.Volume, 0)
			}
			if got := test.terms.GetVolume(); got != test.want {
				t.Errorf("GetVolume() got: %v | want: %v", got, test.want)
			}
			if test.terms.Volume != test.want { // must be the same value with test.want after called GetVolume()
				t.Errorf("ProviderTerms.Volume is: %v | want: %v", test.terms.Volume, test.want)
			}
		})
	}
}

func Test_ProviderTerms_Increase(t *testing.T) {
	t.Parallel()

	terms := mockProviderTerms()

	// the volume must to be zeroed
	terms.Volume = 0
	// prolong expire of terms
	terms.ExpiredAt += time.Timestamp(terms.ProlongDuration)
	if terms.PriceAutoUpdate != 0 {
		terms.Price += terms.PriceAutoUpdate // up the price
	}
	if terms.QoSAutoUpdate.UploadMbps != 0 && terms.QoS.UploadMbps > terms.QoSAutoUpdate.UploadMbps {
		terms.QoS.UploadMbps -= terms.QoSAutoUpdate.UploadMbps // down the qos of upload mbps
	}
	if terms.QoSAutoUpdate.DownloadMbps != 0 && terms.QoS.DownloadMbps > terms.QoSAutoUpdate.DownloadMbps {
		terms.QoS.DownloadMbps -= terms.QoSAutoUpdate.DownloadMbps // down the qos of download mbps
	}

	tests := [1]struct {
		name  string
		terms ProviderTerms
		want  ProviderTerms
	}{
		{
			name:  "OK",
			terms: mockProviderTerms(),
			want:  terms,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			test.terms.Increase()
			if !reflect.DeepEqual(test.terms, test.want) {
				t.Errorf("Increase() got: %#v | want: %#v", test.terms, test.want)
			}
		})
	}
}

func Test_ProviderTerms_Validate(t *testing.T) {
	t.Parallel()

	termsZeroExpiredAt := mockProviderTerms()
	termsZeroExpiredAt.ExpiredAt = 0

	termsZeroQoSUploadMbps := mockProviderTerms()
	termsZeroQoSUploadMbps.QoS.UploadMbps = 0

	termsZeroQoSDownloadMbps := mockProviderTerms()
	termsZeroQoSDownloadMbps.QoS.DownloadMbps = 0

	tests := [4]struct {
		name  string
		terms ProviderTerms
		error bool
	}{
		{
			name:  "OK",
			terms: mockProviderTerms(),
			error: false,
		},
		{
			name:  "Zero_Expired_At_ERR",
			terms: termsZeroExpiredAt,
			error: true,
		},
		{
			name:  "Zero_QoS_Upload_Mbps_ERR",
			terms: termsZeroQoSUploadMbps,
			error: true,
		},
		{
			name:  "Zero_QoS_Download_Mbps_ERR",
			terms: termsZeroQoSDownloadMbps,
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.terms.Validate(); (err != nil) != test.error {
				t.Errorf("Validate() error: %v | want: %v", err, test.error)
			}
		})
	}
}
