package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_TokenPoolTransfer_Decode(t *testing.T) {
	t.Parallel()

	resp := mockTokenPoolTransfer()
	blob, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [2]struct {
		name  string
		blob  []byte
		want  TokenPoolTransfer
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  resp,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"), // invalid json
			want:  TokenPoolTransfer{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := TokenPoolTransfer{}
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

func Test_TokenPoolTransfer_Encode(t *testing.T) {
	t.Parallel()

	resp := mockTokenPoolTransfer()
	blob, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name string
		resp TokenPoolTransfer
		want []byte
	}{
		{
			name: "OK",
			resp: resp,
			want: blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.resp.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}
