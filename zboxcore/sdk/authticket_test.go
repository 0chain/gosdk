package sdk

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitAuthTicket(t *testing.T) {
	type args struct {
		authTicket string
	}
	tests := []struct {
		name string
		args args
		want *AuthTicket
	}{
		{
			"Test_Coverage",
			args{"eyJjbGllbnRfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4Iiwib3duZXJfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4IiwiYWxsb2NhdGlvbl9pZCI6IjY5ZmU1MDM1NTFlZWE1NTU5YzkyNzEyZGZmYzkzMmQ4Y2ZlY2Q4YWU2NDFiMmYyNDJkYjI5ODg3ZTljZTYxOGYiLCJmaWxlX3BhdGhfaGFzaCI6ImM4ODRhYmIzMmFhMDM1N2UyNTQxYjY4M2Y2ZTUyYmZhYjkxNDNkMzNiOTY4OTc3Y2Y2YmEzMWI0M2U4MzI2OTciLCJmaWxlX25hbWUiOiIxLnR4dCIsInJlZmVyZW5jZV90eXBlIjoiZiIsImV4cGlyYXRpb24iOjE2MjQ5OTE3NDcsInRpbWVzdGFtcCI6MTYxNzIxNTc0NywicmVfZW5jcnlwdGlvbl9rZXkiOiIiLCJzaWduYXR1cmUiOiI1Mjk3Y2UyYzVlNzU1NTFhMmJmNWEzMmQ3YmU2MzM4N2U5NzIxZTM2N2QzMDc5ZTI1ZmViZDFkMmIxMWE2NzIwIn0="},
			&AuthTicket{"eyJjbGllbnRfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4Iiwib3duZXJfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4IiwiYWxsb2NhdGlvbl9pZCI6IjY5ZmU1MDM1NTFlZWE1NTU5YzkyNzEyZGZmYzkzMmQ4Y2ZlY2Q4YWU2NDFiMmYyNDJkYjI5ODg3ZTljZTYxOGYiLCJmaWxlX3BhdGhfaGFzaCI6ImM4ODRhYmIzMmFhMDM1N2UyNTQxYjY4M2Y2ZTUyYmZhYjkxNDNkMzNiOTY4OTc3Y2Y2YmEzMWI0M2U4MzI2OTciLCJmaWxlX25hbWUiOiIxLnR4dCIsInJlZmVyZW5jZV90eXBlIjoiZiIsImV4cGlyYXRpb24iOjE2MjQ5OTE3NDcsInRpbWVzdGFtcCI6MTYxNzIxNTc0NywicmVfZW5jcnlwdGlvbl9rZXkiOiIiLCJzaWduYXR1cmUiOiI1Mjk3Y2UyYzVlNzU1NTFhMmJmNWEzMmQ3YmU2MzM4N2U5NzIxZTM2N2QzMDc5ZTI1ZmViZDFkMmIxMWE2NzIwIn0="},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, InitAuthTicket(tt.args.authTicket))
		})
	}
}

func TestAuthTicket_IsDir(t *testing.T) {
	type fields struct {
		b64Ticket string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			"Test_Base64_Decode_Failed",
			fields{b64Ticket: "\\\\\\"},
			false,
			true,
		},
		{
			"Test_Unmarshal_JSON_Auth_Ticket_Failed",
			fields{b64Ticket: "dGhpcyBpcyBub3QganNvbiBzY2hlbWE="},
			false,
			true,
		},
		{
			"Test_Non_Directory_Success",
			fields{b64Ticket: "eyJjbGllbnRfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4Iiwib3duZXJfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4IiwiYWxsb2NhdGlvbl9pZCI6IjY5ZmU1MDM1NTFlZWE1NTU5YzkyNzEyZGZmYzkzMmQ4Y2ZlY2Q4YWU2NDFiMmYyNDJkYjI5ODg3ZTljZTYxOGYiLCJmaWxlX3BhdGhfaGFzaCI6ImM4ODRhYmIzMmFhMDM1N2UyNTQxYjY4M2Y2ZTUyYmZhYjkxNDNkMzNiOTY4OTc3Y2Y2YmEzMWI0M2U4MzI2OTciLCJmaWxlX25hbWUiOiIxLnR4dCIsInJlZmVyZW5jZV90eXBlIjoiZiIsImV4cGlyYXRpb24iOjE2MjQ5OTE3NDcsInRpbWVzdGFtcCI6MTYxNzIxNTc0NywicmVfZW5jcnlwdGlvbl9rZXkiOiIiLCJzaWduYXR1cmUiOiI1Mjk3Y2UyYzVlNzU1NTFhMmJmNWEzMmQ3YmU2MzM4N2U5NzIxZTM2N2QzMDc5ZTI1ZmViZDFkMmIxMWE2NzIwIn0="},
			false,
			false,
		},
		{
			"Test_Is_Directory_Success",
			fields{b64Ticket: "eyJjbGllbnRfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4Iiwib3duZXJfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4IiwiYWxsb2NhdGlvbl9pZCI6IjY5ZmU1MDM1NTFlZWE1NTU5YzkyNzEyZGZmYzkzMmQ4Y2ZlY2Q4YWU2NDFiMmYyNDJkYjI5ODg3ZTljZTYxOGYiLCJmaWxlX3BhdGhfaGFzaCI6ImM4ODRhYmIzMmFhMDM1N2UyNTQxYjY4M2Y2ZTUyYmZhYjkxNDNkMzNiOTY4OTc3Y2Y2YmEzMWI0M2U4MzI2OTciLCJmaWxlX25hbWUiOiIxLnR4dCIsInJlZmVyZW5jZV90eXBlIjoiZCIsImV4cGlyYXRpb24iOjE2MjQ5OTI1MDIsInRpbWVzdGFtcCI6MTYxNzIxNjUwMiwicmVfZW5jcnlwdGlvbl9rZXkiOiIiLCJzaWduYXR1cmUiOiJkOTMyZjE2ZjRkYWI1MTBjMjg2YTJiZWIwZTM2NWNkMDg1NzdlZTFkYjQ2YjgxMDZlNWY0N2JkZDZkZGZlYzBkIn0="},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := &AuthTicket{
				b64Ticket: tt.fields.b64Ticket,
			}
			got, err := at.IsDir()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.want {
				assert.True(t, got)
				return
			}
			assert.False(t, got)
		})
	}
}

func TestAuthTicket_GetFileName(t *testing.T) {
	type fields struct {
		b64Ticket string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			"Test_Base64_Decode_Failed",
			fields{b64Ticket: "\\\\\\"},
			"",
			true,
		},
		{
			"Test_Unmarshal_JSON_Auth_Ticket_Failed",
			fields{b64Ticket: "dGhpcyBpcyBub3QganNvbiBzY2hlbWE="},
			"",
			true,
		},
		{
			"Test_Get_File_Name_Success",
			fields{b64Ticket: "eyJjbGllbnRfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4Iiwib3duZXJfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4IiwiYWxsb2NhdGlvbl9pZCI6IjY5ZmU1MDM1NTFlZWE1NTU5YzkyNzEyZGZmYzkzMmQ4Y2ZlY2Q4YWU2NDFiMmYyNDJkYjI5ODg3ZTljZTYxOGYiLCJmaWxlX3BhdGhfaGFzaCI6ImM4ODRhYmIzMmFhMDM1N2UyNTQxYjY4M2Y2ZTUyYmZhYjkxNDNkMzNiOTY4OTc3Y2Y2YmEzMWI0M2U4MzI2OTciLCJmaWxlX25hbWUiOiIxLnR4dCIsInJlZmVyZW5jZV90eXBlIjoiZiIsImV4cGlyYXRpb24iOjE2MjQ5OTE3NDcsInRpbWVzdGFtcCI6MTYxNzIxNTc0NywicmVfZW5jcnlwdGlvbl9rZXkiOiIiLCJzaWduYXR1cmUiOiI1Mjk3Y2UyYzVlNzU1NTFhMmJmNWEzMmQ3YmU2MzM4N2U5NzIxZTM2N2QzMDc5ZTI1ZmViZDFkMmIxMWE2NzIwIn0="},
			"1.txt",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := &AuthTicket{
				b64Ticket: tt.fields.b64Ticket,
			}
			got, err := at.GetFileName()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthTicket_GetLookupHash(t *testing.T) {
	type fields struct {
		b64Ticket string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			"Test_Base64_Decode_Failed",
			fields{b64Ticket: "\\\\\\"},
			"",
			true,
		},
		{
			"Test_Unmarshal_JSON_Auth_Ticket_Failed",
			fields{b64Ticket: "dGhpcyBpcyBub3QganNvbiBzY2hlbWE="},
			"",
			true,
		},
		{
			"Test_Get_File_Lookup_Hash_Success",
			fields{b64Ticket: "eyJjbGllbnRfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4Iiwib3duZXJfaWQiOiI5YmY0MzBkNmYwODZmMWJkYzJkMjZhZDdhNzA4YTBlNzk1OGFhOWFlMjBlZmJjNjc3ODQ1MDczOWZiMWNhNDY4IiwiYWxsb2NhdGlvbl9pZCI6IjY5ZmU1MDM1NTFlZWE1NTU5YzkyNzEyZGZmYzkzMmQ4Y2ZlY2Q4YWU2NDFiMmYyNDJkYjI5ODg3ZTljZTYxOGYiLCJmaWxlX3BhdGhfaGFzaCI6ImM4ODRhYmIzMmFhMDM1N2UyNTQxYjY4M2Y2ZTUyYmZhYjkxNDNkMzNiOTY4OTc3Y2Y2YmEzMWI0M2U4MzI2OTciLCJmaWxlX25hbWUiOiIxLnR4dCIsInJlZmVyZW5jZV90eXBlIjoiZiIsImV4cGlyYXRpb24iOjE2MjQ5OTE3NDcsInRpbWVzdGFtcCI6MTYxNzIxNTc0NywicmVfZW5jcnlwdGlvbl9rZXkiOiIiLCJzaWduYXR1cmUiOiI1Mjk3Y2UyYzVlNzU1NTFhMmJmNWEzMmQ3YmU2MzM4N2U5NzIxZTM2N2QzMDc5ZTI1ZmViZDFkMmIxMWE2NzIwIn0="},
			"c884abb32aa0357e2541b683f6e52bfab9143d33b968977cf6ba31b43e832697",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := &AuthTicket{
				b64Ticket: tt.fields.b64Ticket,
			}
			got, err := at.GetLookupHash()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
