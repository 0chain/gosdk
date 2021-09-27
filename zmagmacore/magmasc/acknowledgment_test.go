package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_Acknowledgment_Decode(t *testing.T) {
	t.Parallel()

	ackn := mockAcknowledgment()
	blob, err := json.Marshal(ackn)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	acknInvalid := mockAcknowledgment()
	acknInvalid.SessionID = ""
	blobInvalid, err := json.Marshal(acknInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [3]struct {
		name  string
		blob  []byte
		want  *Acknowledgment
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  ackn,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"), // invalid json
			want:  &Acknowledgment{},
			error: true,
		},
		{
			name:  "Invalid_ERR",
			blob:  blobInvalid,
			want:  &Acknowledgment{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := &Acknowledgment{}
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

func Test_Acknowledgment_Encode(t *testing.T) {
	t.Parallel()

	ackn := mockAcknowledgment()
	blob, err := json.Marshal(ackn)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name string
		ackn *Acknowledgment
		want []byte
	}{
		{
			name: "OK",
			ackn: ackn,
			want: blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.ackn.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_Acknowledgment_Key(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		ackn := mockAcknowledgment()
		want := []byte(AcknowledgmentPrefix + ackn.SessionID)

		if got := ackn.Key(); !reflect.DeepEqual(got, want) {
			t.Errorf("Key() got: %v | want: %v", string(got), string(want))
		}
	})
}

func Test_Acknowledgment_Validate(t *testing.T) {
	t.Parallel()

	acknEmptySessionID := mockAcknowledgment()
	acknEmptySessionID.SessionID = ""

	acknEmptyAccessPointID := mockAcknowledgment()
	acknEmptyAccessPointID.AccessPointID = ""

	acknEmptyConsumerExtID := mockAcknowledgment()
	acknEmptyConsumerExtID.Consumer.ExtID = ""

	acknEmptyProviderExtID := mockAcknowledgment()
	acknEmptyProviderExtID.Provider.ExtID = ""

	tests := [5]struct {
		name  string
		ackn  *Acknowledgment
		error bool
	}{
		{
			name:  "OK",
			ackn:  mockAcknowledgment(),
			error: false,
		},
		{
			name:  "Empty_Session_ID",
			ackn:  acknEmptySessionID,
			error: true,
		},
		{
			name:  "Empty_Access_Point_ID",
			ackn:  acknEmptyAccessPointID,
			error: true,
		},
		{
			name:  "Empty_Consumer_Ext_ID",
			ackn:  acknEmptyConsumerExtID,
			error: true,
		},
		{
			name:  "Empty_Provider_Txt_ID",
			ackn:  acknEmptyProviderExtID,
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.ackn.Validate(); (err != nil) != test.error {
				t.Errorf("validate() error: %v | want: %v", err, test.error)
			}
		})
	}
}
