package blobberClient

import "testing"

func Test_getGRPCPort(t *testing.T) {
	type args struct {
		port string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "5051 -> 31501",
			args: args{ port: "5051" },
			want: "31501",
		},
		{
			name: "31306 -> 31506",
			args: args{ port: "31306" },
			want: "31506",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGRPCPort(tt.args.port); got != tt.want {
				t.Errorf("getGRPCPort() = %v, want %v", got, tt.want)
			}
		})
	}
}
