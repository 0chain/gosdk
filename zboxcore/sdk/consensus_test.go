package sdk

import (
	"testing"

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
