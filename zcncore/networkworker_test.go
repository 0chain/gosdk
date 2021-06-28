package zcncore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateNetworkDetailsWorker(t *testing.T) {
	t.Run("Update Network Details Worker Stop by user", func(t *testing.T) {
		ctx, cancelFn := context.WithCancel(context.Background())
		go UpdateNetworkDetailsWorker(ctx)
		cancelFn()
	})
	
}
func TestUpdateNetworkDetails(t *testing.T) {
	t.Run("Update Network Details Success", func(t *testing.T) {
		err := UpdateNetworkDetails()
		require.NoError(t, err)
	})
}
func TestUpdateRequired(t *testing.T) {
	t.Run("Update Required return false", func(t *testing.T) {
		_config.chain.Miners = []string{"1", "2"}
		_config.chain.Sharders = []string{"3", "4"}
		check := UpdateRequired(&Network{
			Miners:   []string{"1", "2"},
			Sharders: []string{"3", "4"},
		})
		require.Equal(t, false, check)
	})
	t.Run("Update Required return true", func(t *testing.T) {
		_config.chain.Miners = []string{}
		_config.chain.Sharders = []string{"3", "4"}
		check := UpdateRequired(&Network{
			Miners:   []string{"1", "2"},
			Sharders: []string{"3", "4"},
		})
		require.Equal(t, true, check)
	})
	t.Run("Update Required return true miner not equal case", func(t *testing.T) {
		_config.chain.Miners = []string{"1","3"}
		_config.chain.Sharders = []string{"3", "4"}
		check := UpdateRequired(&Network{
			Miners:   []string{"1", "2"},
			Sharders: []string{"3", "4"},
		})
		require.Equal(t, true, check)
	})
}
