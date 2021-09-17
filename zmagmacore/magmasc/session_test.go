package magmasc

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_Session_Decode(t *testing.T) {
	t.Parallel()

	session := mockSession()
	blob, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	sessionInvalid := mockSession()
	sessionInvalid.SessionID = ""
	blobInvalid, err := json.Marshal(sessionInvalid)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [3]struct {
		name  string
		blob  []byte
		want  *Session
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  session,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"), // invalid json
			want:  &Session{},
			error: true,
		},
		{
			name:  "Invalid_ERR",
			blob:  blobInvalid,
			want:  &Session{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := &Session{}
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

func Test_Session_Encode(t *testing.T) {
	t.Parallel()

	session := mockSession()
	blob, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name    string
		session *Session
		want    []byte
	}{
		{
			name:    "OK",
			session: session,
			want:    blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.session.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_Session_Key(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		session := mockSession()
		want := []byte(SessionPrefix + session.SessionID)

		if got := session.Key(); !reflect.DeepEqual(got, want) {
			t.Errorf("Key() got: %v | want: %v", string(got), string(want))
		}
	})
}

func Test_Session_Validate(t *testing.T) {
	t.Parallel()

	sessionEmptySessionID := mockSession()
	sessionEmptySessionID.SessionID = ""

	sessionEmptyAccessPointID := mockSession()
	sessionEmptyAccessPointID.AccessPoint.ID = ""

	sessionEmptyConsumerExtID := mockSession()
	sessionEmptyConsumerExtID.Consumer.ExtID = ""

	sessionEmptyProviderExtID := mockSession()
	sessionEmptyProviderExtID.Provider.ExtID = ""

	tests := [5]struct {
		name    string
		session *Session
		error   bool
	}{
		{
			name:    "OK",
			session: mockSession(),
			error:   false,
		},
		{
			name:    "Empty_Session_ID",
			session: sessionEmptySessionID,
			error:   true,
		},
		{
			name:    "Empty_Access_Point_ID",
			session: sessionEmptyAccessPointID,
			error:   true,
		},
		{
			name:    "Empty_Consumer_Ext_ID",
			session: sessionEmptyConsumerExtID,
			error:   true,
		},
		{
			name:    "Empty_Provider_Txt_ID",
			session: sessionEmptyProviderExtID,
			error:   true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.session.Validate(); (err != nil) != test.error {
				t.Errorf("validate() error: %v | want: %v", err, test.error)
			}
		})
	}
}
