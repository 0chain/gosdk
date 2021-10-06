package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zmagmacore/magmasc/pb"
)

func Test_DataMarker_Decode(t *testing.T) {
	t.Parallel()

	dataMarker := mockDataMarker()
	blob, err := json.Marshal(dataMarker)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	dataMarkerInvalid := mockDataMarker()
	dataMarkerInvalid.UserId = ""
	blobInvalid, err := json.Marshal(dataMarkerInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [3]struct {
		name  string
		blob  []byte
		want  *DataMarker
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  dataMarker,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"),
			want:  &DataMarker{},
			error: true,
		},
		{
			name:  "User_ID_Invalid_ERR",
			blob:  blobInvalid,
			want:  &DataMarker{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := &DataMarker{}
			if err := got.Decode(test.blob); (err != nil) != test.error {
				t.Errorf("Decode() error: %v | want: %v", err, nil)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Decode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_DataMarker_Encode(t *testing.T) {
	t.Parallel()

	dataMarker := mockDataMarker()
	blob, err := json.Marshal(dataMarker)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name string
		data *DataMarker
		want []byte
	}{
		{
			name: "OK",
			data: dataMarker,
			want: blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.data.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}

}

func Test_DataMarker_Sign(t *testing.T) {
	t.Parallel()

	schemeType := "bls0chain"
	scheme := zcncrypto.NewSignatureScheme(schemeType)
	if _, err := scheme.GenerateKeys(); err != nil {
		t.Fatalf("zcncrypto.GenerateKeys() error: %v | want: %v", err, nil)
	}

	type (
		args struct {
			scheme     zcncrypto.SignatureScheme
			schemeType string
		}
	)
	tests := []struct {
		name       string
		DataMarker *DataMarker
		args       args
		wantErr    bool
	}{
		{
			name:       "OK",
			DataMarker: mockDataMarker(),
			args: args{
				scheme:     scheme,
				schemeType: schemeType,
			},
			wantErr: false,
		},
		{
			name:       "Empty_Private_Key_ERR",
			DataMarker: mockDataMarker(),
			args: args{
				scheme:     zcncrypto.NewSignatureScheme(schemeType),
				schemeType: schemeType,
			},
			wantErr: true,
		},
	}
	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.DataMarker.Sign(test.args.scheme, test.args.schemeType); (err != nil) != test.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func Test_DataMarker_Validate(t *testing.T) {
	t.Parallel()

	dataMarker := mockDataMarker()

	tests := []struct {
		name  string
		data  *DataMarker
		error bool
	}{
		{
			name: "Regular_Marker_OK",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					DataUsage: mockDataUsage(),
				},
			},
			error: false,
		},
		{
			name: "Nil_Data_Usage_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					DataUsage: nil,
				},
			},
			error: true,
		},
		{
			name: "Empty_SessionID_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					DataUsage: &pb.DataUsage{
						SessionId: "",
					},
				},
			},
			error: true,
		},
		{
			name:  "QoS_Marker_OK",
			data:  mockDataMarker(),
			error: false,
		},
		{
			name: "QoS_Marker_Invalid_UserID_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    "",
					DataUsage: dataMarker.DataUsage,
					Qos:       dataMarker.Qos,
					PublicKey: dataMarker.PublicKey,
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_Nil_QoS_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: mockDataUsage(),
					Qos:       nil,
					PublicKey: dataMarker.PublicKey,
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_Invalid_QoS_DownloadMbps_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: mockDataUsage(),
					Qos: &pb.QoS{
						DownloadMbps: 0,
						UploadMbps:   dataMarker.Qos.UploadMbps,
						Latency:      dataMarker.Qos.Latency,
					},
					PublicKey: dataMarker.PublicKey,
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_Invalid_QoS_UploadMbps_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: mockDataUsage(),
					Qos: &pb.QoS{
						DownloadMbps: dataMarker.Qos.DownloadMbps,
						UploadMbps:   0,
						Latency:      dataMarker.Qos.Latency,
					},
					PublicKey: dataMarker.PublicKey,
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_Invalid_QoS_Latency_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: mockDataUsage(),
					Qos: &pb.QoS{
						DownloadMbps: dataMarker.Qos.DownloadMbps,
						UploadMbps:   dataMarker.Qos.UploadMbps,
						Latency:      0,
					},
					PublicKey: dataMarker.PublicKey,
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_Invalid_Public_Key_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: mockDataUsage(),
					Qos: &pb.QoS{
						DownloadMbps: dataMarker.Qos.DownloadMbps,
						UploadMbps:   dataMarker.Qos.UploadMbps,
						Latency:      dataMarker.Qos.Latency,
					},
					PublicKey: "",
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_Invalid_SigScheme_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: mockDataUsage(),
					Qos: &pb.QoS{
						DownloadMbps: dataMarker.Qos.DownloadMbps,
						UploadMbps:   dataMarker.Qos.UploadMbps,
						Latency:      dataMarker.Qos.Latency,
					},
					PublicKey: dataMarker.PublicKey,
					SigScheme: "",
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_Invalid_Signature_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: mockDataUsage(),
					Qos: &pb.QoS{
						DownloadMbps: dataMarker.Qos.DownloadMbps,
						UploadMbps:   dataMarker.Qos.UploadMbps,
						Latency:      dataMarker.Qos.Latency,
					},
					PublicKey: dataMarker.PublicKey,
					SigScheme: dataMarker.SigScheme,
					Signature: "",
				},
			},
			error: true,
		},
		{
			name: "QoS_Marker_UserID_And_Public_Key_Mismatching_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    "invalid_user_id",
					DataUsage: mockDataUsage(),
					Qos: &pb.QoS{
						DownloadMbps: dataMarker.Qos.DownloadMbps,
						UploadMbps:   dataMarker.Qos.UploadMbps,
						Latency:      dataMarker.Qos.Latency,
					},
					PublicKey: dataMarker.PublicKey,
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.data.Validate(); (err != nil) != test.error {
				t.Errorf("validate() error: %v | want: %v", err, test.error)
			}
		})
	}
}

func Test_DataMarker_Verify(t *testing.T) {
	t.Parallel()

	var (
		dataMarker = mockDataMarker()

		schemeType = "bls0chain"
		scheme     = zcncrypto.NewSignatureScheme(schemeType)
	)
	_, err := scheme.GenerateKeys()
	if err != nil {
		t.Fatalf("SignatureScheme.GenerateKeys() error: %v | want: %v", err, nil)
	}
	if err := dataMarker.Sign(scheme, schemeType); err != nil {
		t.Fatalf("DataMarker.Sign() error: %v | want: %v", err, nil)
	}

	anotherScheme := zcncrypto.NewSignatureScheme(schemeType)
	_, err = anotherScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("SignatureScheme.GenerateKeys() error: %v | want: %v", err, nil)
	}

	tests := []struct {
		name   string
		data   *DataMarker
		scheme zcncrypto.SignatureScheme
		want   bool
		error  bool
	}{
		{
			name:  "TRUE",
			data:  dataMarker,
			want:  true,
			error: false,
		},
		{
			name: "Wrong_Scheme_FALSE",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: dataMarker.DataUsage,
					Qos:       dataMarker.Qos,
					PublicKey: anotherScheme.GetPublicKey(),
					SigScheme: schemeType,
					Signature: dataMarker.Signature,
				},
			},
			want:  false,
			error: false,
		},
		{
			name: "Empty_Public_Key_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: dataMarker.DataUsage,
					Qos:       dataMarker.Qos,
					PublicKey: "",
					SigScheme: dataMarker.SigScheme,
					Signature: dataMarker.Signature,
				},
			},
			want:  false,
			error: true,
		},
		{
			name: "Unsupported_scheme_ERR",
			data: &DataMarker{
				DataMarker: &pb.DataMarker{
					UserId:    dataMarker.UserId,
					DataUsage: dataMarker.DataUsage,
					Qos:       dataMarker.Qos,
					PublicKey: dataMarker.PublicKey,
					SigScheme: "unsupported",
					Signature: dataMarker.Signature,
				},
			},
			want:  false,
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ok, err := test.data.Verify()
			if (err != nil) != test.error {
				t.Errorf("Verify() error: %v | want: %v", err, test.error)
			}
			if ok != test.want {
				t.Errorf("Verify() got: %v | want: %v", ok, test.want)
			}
		})
	}
}
