package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/0chain/gosdk/core/zcncrypto"
)

func Test_UserDataMarker_Decode(t *testing.T) {
	t.Parallel()

	dataMarker := mockUserDataMarker()
	blob, err := json.Marshal(dataMarker)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	dataMarkerInvalid := mockUserDataMarker()
	dataMarkerInvalid.UserID = ""
	blobInvalid, err := json.Marshal(dataMarkerInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [3]struct {
		name  string
		blob  []byte
		want  *UserDataMarker
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
			want:  &UserDataMarker{},
			error: true,
		},
		{
			name:  "User_ID_Invalid_ERR",
			blob:  blobInvalid,
			want:  &UserDataMarker{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := &UserDataMarker{}
			if err := got.Decode(test.blob); (err != nil) != test.error {
				t.Errorf("Decode() error: %v | want: %v", err, nil)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Decode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_UserDataMarker_Encode(t *testing.T) {
	t.Parallel()

	dataMarker := mockUserDataMarker()
	blob, err := json.Marshal(dataMarker)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name string
		data *UserDataMarker
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

func Test_UserDataMarker_Validate(t *testing.T) {
	t.Parallel()

	dataMarkerEmptyUserID := mockUserDataMarker()
	dataMarkerEmptyUserID.UserID = ""

	tests := [2]struct {
		name  string
		data  *UserDataMarker
		error bool
	}{
		{
			name:  "OK",
			data:  mockUserDataMarker(),
			error: false,
		},
		{
			name:  "Empty_User_ID_ERR",
			data:  dataMarkerEmptyUserID,
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

func Test_UserDataMarker_Verify(t *testing.T) {
	t.Parallel()

	dataMarker := mockUserDataMarker()
	scheme := zcncrypto.NewSignatureScheme("bls0chain")
	_, err := scheme.GenerateKeys()
	if err != nil {
		t.Fatalf("SignatureScheme.GenerateKeys() error: %v | want: %v", err, nil)
	}
	if err := dataMarker.Sign(scheme); err != nil {
		t.Fatalf("UserDataMarker.Sign() error: %v | want: %v", err, nil)
	}

	anotherScheme := zcncrypto.NewSignatureScheme("bls0chain")
	_, err = anotherScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("SignatureScheme.GenerateKeys() error: %v | want: %v", err, nil)
	}

	tests := [2]struct {
		name   string
		data   *UserDataMarker
		scheme zcncrypto.SignatureScheme
		want   bool
		error  bool
	}{
		{
			name:   "TRUE",
			data:   dataMarker,
			scheme: scheme,
			want:   true,
			error:  false,
		},
		{
			name:   "Wrong_Scheme_FALSE",
			data:   dataMarker,
			scheme: anotherScheme,
			want:   false,
			error:  false,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ok, err := test.data.Verify(test.scheme)
			if (err != nil) != test.error {
				t.Errorf("Verify() error: %v | want: %v", err, test.error)
			}
			if ok != test.want {
				t.Errorf("Verify() got: %v | want: %v", ok, test.want)
			}
		})
	}
}
