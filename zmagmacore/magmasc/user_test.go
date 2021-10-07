package magmasc

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func Test_User_Decode(t *testing.T) {
	t.Parallel()

	user := mockUser()
	blob, err := json.Marshal(user.User)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	userIDInvalid := mockUser()
	userIDInvalid.ID = "" // invalid user's id
	userIDBlobInvalid, err := json.Marshal(userIDInvalid.User)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	consumerIDInvalid := mockUser()
	consumerIDInvalid.ConsumerID = "" // invalid consumer's id
	consumerIDBlobInvalid, err := json.Marshal(consumerIDInvalid.User)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [4]struct {
		name  string
		blob  []byte
		want  *User
		error bool
	}{
		{
			name:  "OK",
			blob:  blob,
			want:  user,
			error: false,
		},
		{
			name:  "Decode_ERR",
			blob:  []byte(":"), // invalid json
			want:  &User{},
			error: true,
		},
		{
			name:  "User_ID_Invalid_ERR",
			blob:  userIDBlobInvalid,
			want:  &User{},
			error: true,
		},
		{
			name:  "Consumer_ID_Invalid_ERR",
			blob:  consumerIDBlobInvalid,
			want:  &User{},
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := &User{}
			if err := got.Decode(test.blob); (err != nil) != test.error {
				t.Errorf("Decode() error: %v | want: %v", err, nil)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Decode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_User_Encode(t *testing.T) {
	t.Parallel()

	user := mockUser()
	blob, err := json.Marshal(user.User)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name string
		user *User
		want []byte
	}{
		{
			name: "OK",
			user: user,
			want: blob,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.user.Encode(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Encode() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}

func Test_User_Validate(t *testing.T) {
	t.Parallel()

	userEmptyID := mockUser()
	userEmptyID.ID = "" // invalid user's id

	consumerEmptyID := mockUser()
	consumerEmptyID.ConsumerID = "" // invalid consumer's id

	tests := [3]struct {
		name  string
		cons  *User
		error bool
	}{
		{
			name:  "OK",
			cons:  mockUser(),
			error: false,
		},
		{
			name:  "Empty_User_ID_ERR",
			cons:  userEmptyID,
			error: true,
		},
		{
			name:  "Empty_Consumer_ID_ERR",
			cons:  consumerEmptyID,
			error: true,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if err := test.cons.Validate(); (err != nil) != test.error {
				t.Errorf("Validate() error: %v | want: %v", err, test.error)
			}
		})
	}
}

func Test_User_ReadYAML(t *testing.T) {
	t.Parallel()

	var (
		buf = bytes.NewBuffer(nil)
		enc = yaml.NewEncoder(buf)

		user = mockUser()
	)

	err := enc.Encode(user)
	if err != nil {
		t.Fatalf("yaml Encode() error: %v | want: %v", err, nil)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	err = os.WriteFile(path, buf.Bytes(), 0777)
	if err != nil {
		t.Fatalf("os.WriteFile() error: %v | want: %v", err, nil)
	}

	tests := [1]struct {
		name  string
		want  *User
		error bool
	}{
		{
			name:  "OK",
			want:  user,
			error: false,
		},
	}

	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := &User{}
			if err := got.ReadYAML(path); (err != nil) != test.error {
				t.Errorf("ReadYAML() error: %v | want: %v", err, nil)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ReadYAML() got: %#v | want: %#v", got, test.want)
			}
		})
	}
}
