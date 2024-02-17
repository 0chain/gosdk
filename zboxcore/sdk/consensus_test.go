package sdk

import (
	"sync"
	"testing"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/minio/sha256-simd"
	"github.com/stretchr/testify/require"
)

func TestConsensus_isConsensusOk(t *testing.T) {
	type fields struct {
		consensus              int
		consensusThresh        int
		fullconsensus          int
		consensusRequiredForOk int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"Test_Is_Consensus_OK_True",
			fields{
				consensus:              3,
				consensusThresh:        2,
				fullconsensus:          4,
				consensusRequiredForOk: 60,
			},
			true,
		},
		{
			"Test_Is_Consensus_OK_False",
			fields{
				consensus:              2,
				consensusThresh:        3,
				fullconsensus:          4,
				consensusRequiredForOk: 60,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Consensus{
				RWMutex:         &sync.RWMutex{},
				consensus:       tt.fields.consensus,
				consensusThresh: tt.fields.consensusThresh,
				fullconsensus:   tt.fields.fullconsensus,
			}
			got := req.isConsensusOk()
			require := require.New(t)
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(got)
		})
	}
}

func TestHash(t *testing.T) {
	hasher := sha256.New()
	b := []byte("hello")
	res := encryption.ShaHash(b)
	hasher.Write(b[:2])
	hasher.Write(b[2:])
	require.Equal(t, res, hasher.Sum(nil))
}
