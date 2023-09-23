package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeHolder_Success(t *testing.T) {
	type fields struct {
		nodes     []string
		consensus int
	}
	type args struct {
		id string
	}
	type res struct {
		res []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{name: "init", fields: struct {
			nodes     []string
			consensus int
		}{nodes: []string{"1", "2", "3", "4", "5"}, consensus: 5}, args: struct{ id string }{id: "1"},
			res: struct{ res []string }{res: []string{"1", "2", "3", "4", "5"}}},
		{name: "pull up", fields: struct {
			nodes     []string
			consensus int
		}{nodes: []string{"1", "2", "3", "4", "5"}, consensus: 5}, args: struct{ id string }{id: "5"},
			res: struct{ res []string }{res: []string{"5", "1", "2", "3", "4"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHolder(tt.fields.nodes, tt.fields.consensus)
			h.Success(tt.args.id)

			assert.Equal(t, tt.res.res, h.Healthy())
		})
	}
}
