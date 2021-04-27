package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsensus_isConsensusMin(t *testing.T) {
	type fields struct {
		consensus       float32
		consensusThresh float32
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
		{
			"Test_IsConsensusMin_True",
			fields{
				consensus:       1,
				consensusThresh: 1,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Consensus{
				consensus:       1,
				consensusThresh: 1,
			}
			got := req.isConsensusMin()
			assertion := assert.New(t)
			var check = assertion.False
			if tt.want {
				check = assertion.True
			}
			check(got)
		})
	}
}
