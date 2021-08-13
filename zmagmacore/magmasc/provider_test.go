package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_Provider_Decode(t *testing.T) {
	t.Parallel()

	prov := mockProvider()
	blob, err := json.Marshal(prov)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	provInvalid := mockProvider()
	provInvalid.ExtID = ""
	extIDBlobInvalid, err := json.Marshal(provInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	provInvalid = mockProvider()
	provInvalid.Terms.QoS.UploadMbps = -0.1
	UploadMbpsBlobInvalid, err := json.Marshal(provInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	provInvalid = mockProvider()
	provInvalid.Terms.QoS.DownloadMbps = -0.1
	DownloadMbpsBlobInvalid, err := json.Marshal(provInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [5]struct {
		name  string
		blob  []byte
		want  *Provider
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  prov,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"), // invalid json
			want:  &Provider{},
			error: true,
		},
		{
			name:  "Ext_ID_Invalid_ERR",
			blob:  extIDBlobInvalid,
			want:  &Provider{},
			error: true,
		},
		{
			name:  "QoS_Upload_Mbps_Invalid_ERR",
			blob:  UploadMbpsBlobInvalid,
			want:  &Provider{},
			error: true,
		},
		{
			name:  "QoS_Download_Mbps_Invalid_ERR",
			blob:  DownloadMbpsBlobInvalid,
			want:  &Provider{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := &Provider{}
			if err := got.Decode(test.blob); (err != nil) != test.error {
				t.Errorf("Decode() error: %v | want: %v", err, nil)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Decode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_Provider_Encode(t *testing.T) {
	t.Parallel()

	prov := mockProvider()
	blob, err := json.Marshal(prov)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name string
		prov *Provider
		want []byte
	}{
		{
			name: "OK",
			prov: prov,
			want: blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.prov.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_Provider_GetType(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		prov := Provider{}
		if got := prov.GetType(); got != providerType {
			t.Errorf("GetType() got: %v | want: %v", got, providerType)
		}
	})
}

func Test_Provider_Validate(t *testing.T) {
	t.Parallel()

	provEmptyExtID := mockProvider()
	provEmptyExtID.ExtID = ""

	provNegTermsQoSUploadMbps := mockProvider()
	provNegTermsQoSUploadMbps.Terms.QoS.UploadMbps = -0.1

	provNegTermsQoSDownloadMbps := mockProvider()
	provNegTermsQoSDownloadMbps.Terms.QoS.DownloadMbps = -0.1

	tests := [4]struct {
		name  string
		prov  *Provider
		error bool
	}{
		{
			name:  "OK",
			prov:  mockProvider(),
			error: false,
		},
		{
			name:  "Empty_Ext_ID_ERR",
			prov:  provEmptyExtID,
			error: true,
		},
		{
			name:  "Neg_Terms_QoS_Upload_Mbps_ERR",
			prov:  provNegTermsQoSUploadMbps,
			error: true,
		},
		{
			name:  "Neg_Terms_QoS_Download_Mbps_ERR",
			prov:  provNegTermsQoSDownloadMbps,
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.prov.Validate(); (err != nil) != test.error {
				t.Errorf("validate() error: %v | want: %v", err, test.error)
			}
		})
	}
}
