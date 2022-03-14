package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsensus_isConsensusMin(t *testing.T) {
	type fields struct {
		consensus       float32
		consensusThresh float32
		fullconsensus   float32
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"Test_Is_Consensus_Min_True",
			fields{
				consensus:       2,
				consensusThresh: 50,
				fullconsensus:   4,
			},
			true,
		},
		{
			"Test_Is_Consensus_Min_False",
			fields{
				consensus:       1,
				consensusThresh: 50,
				fullconsensus:   4,
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
			got := req.isConsensusMin()
			require := require.New(t)
			var check = require.False
			if tt.want {
				check = require.True
			}
			check(got)
		})
	}
}

func TestConsensus_isConsensusOk(t *testing.T) {
	type fields struct {
		consensus              float32
		consensusThresh        float32
		fullconsensus          float32
		consensusRequiredForOk float32
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
				consensusThresh:        50,
				fullconsensus:          4,
				consensusRequiredForOk: 60,
			},
			true,
		},
		{
			"Test_Is_Consensus_OK_False",
			fields{
				consensus:              2,
				consensusThresh:        50,
				fullconsensus:          4,
				consensusRequiredForOk: 60,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Consensus{
				consensus:              tt.fields.consensus,
				consensusThresh:        tt.fields.consensusThresh,
				fullconsensus:          tt.fields.fullconsensus,
				consensusRequiredForOk: tt.fields.consensusRequiredForOk,
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

func TestConsensus_getConsensusRequiredForOk(t *testing.T) {
	type fields struct {
		consensusThresh        float32
		consensusRequiredForOk float32
	}
	tests := []struct {
		name   string
		fields fields
		want   float32
	}{
		{
			"Test_Result_Get_Consensus_Required_For_Ok",
			fields{
				consensusThresh:        50,
				consensusRequiredForOk: 60,
			},
			60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requires := require.New(t)
			req := &Consensus{
				consensusThresh: tt.fields.consensusThresh,
			}
			got := req.getConsensusRequiredForOk()
			requires.Equal(got, tt.want)
		})
	}
}

func TestConsensus_getConsensusRate(t *testing.T) {
	type fields struct {
		consensus     float32
		fullconsensus float32
	}
	tests := []struct {
		name   string
		fields fields
		want   float32
	}{
		{
			"Test_Result_Get_Consensus_Rate",
			fields{
				consensus:     2,
				fullconsensus: 4,
			},
			50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requires := require.New(t)
			req := &Consensus{
				consensus:     tt.fields.consensus,
				fullconsensus: tt.fields.fullconsensus,
			}
			got := req.getConsensusRate()
			requires.Equal(got, tt.want)
		})
	}
}
