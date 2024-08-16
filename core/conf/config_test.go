package conf

import (
	"testing"

	"github.com/0chain/gosdk/core/conf/mocks"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {

	var mockDefaultReader = func() Reader {
		reader := &mocks.Reader{}
		reader.On("GetString", "block_worker").Return("http://127.0.0.1:9091/dns")
		reader.On("GetString", "zauth.server").Return("http://127.0.0.1:8090/")
		reader.On("GetInt", "min_submit").Return(0)
		reader.On("GetInt", "min_confirmation").Return(0)
		reader.On("GetInt", "max_txn_query").Return(0)
		reader.On("GetInt", "query_sleep_time").Return(0)
		reader.On("GetInt", "confirmation_chain_length").Return(0)
		reader.On("GetStringSlice", "preferred_blobbers").Return(nil)
		reader.On("GetString", "signature_scheme").Return("")
		reader.On("GetString", "chain_id").Return("")
		reader.On("GetString", "verify_optimistic").Return("true")
		reader.On("GetInt", "sharder_consensous").Return(0)

		return reader

	}

	tests := []struct {
		name        string
		exceptedErr error

		setup func(*testing.T) Reader
		run   func(*require.Assertions, Config)
	}{
		{
			name:        "Test_Config_Invalid_BlockWorker",
			exceptedErr: ErrInvalidValue,
			setup: func(t *testing.T) Reader {

				reader := &mocks.Reader{}
				reader.On("GetString", "block_worker").Return("")
				reader.On("GetString", "zauth.server").Return("")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)
				reader.On("GetString", "signature_scheme").Return("")
				reader.On("GetString", "chain_id").Return("")
				reader.On("GetString", "verify_optimistic").Return("true")
				reader.On("GetInt", "sharder_consensous").Return(0)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {

			},
		},
		{
			name: "Test_Config_BlockWorker",

			setup: func(t *testing.T) Reader {
				return mockDefaultReader()
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal("http://127.0.0.1:9091/dns", cfg.BlockWorker)
			},
		},
		{
			name: "Test_Config_Min_Submit_Less_Than_1",

			setup: func(t *testing.T) Reader {
				return mockDefaultReader()
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(10, cfg.MinSubmit)
			},
		},
		{
			name: "Test_Config_Min_Submit_Greater_Than_100",

			setup: func(t *testing.T) Reader {

				reader := &mocks.Reader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetString", "zauth.server").Return("http://127.0.0.1:8090/")
				reader.On("GetInt", "min_submit").Return(101)
				reader.On("GetInt", "min_confirmation").Return(0)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)
				reader.On("GetString", "signature_scheme").Return("")
				reader.On("GetString", "chain_id").Return("")
				reader.On("GetString", "verify_optimistic").Return("true")
				reader.On("GetInt", "sharder_consensous").Return(0)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(100, cfg.MinSubmit)
			},
		},
		{
			name: "Test_Config_Min_Confirmation_Less_Than_1",

			setup: func(t *testing.T) Reader {
				return mockDefaultReader()
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(10, cfg.MinConfirmation)
			},
		},
		{
			name: "Test_Config_Min_Confirmation_Greater_100",

			setup: func(t *testing.T) Reader {

				reader := &mocks.Reader{}
				reader.On("GetString", "block_worker").Return("https://127.0.0.1:9091/dns")
				reader.On("GetString", "zauth.server").Return("http://127.0.0.1:8090/")
				reader.On("GetInt", "min_submit").Return(0)
				reader.On("GetInt", "min_confirmation").Return(101)
				reader.On("GetInt", "max_txn_query").Return(0)
				reader.On("GetInt", "query_sleep_time").Return(0)
				reader.On("GetInt", "confirmation_chain_length").Return(0)
				reader.On("GetStringSlice", "preferred_blobbers").Return(nil)
				reader.On("GetString", "signature_scheme").Return("")
				reader.On("GetString", "chain_id").Return("")
				reader.On("GetString", "verify_optimistic").Return("false")
				reader.On("GetInt", "sharder_consensous").Return(0)

				return reader
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(100, cfg.MinConfirmation)
			},
		}, {
			name: "Test_Config_Nax_Txn_Query_Less_Than_1",

			setup: func(t *testing.T) Reader {

				return mockDefaultReader()
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(5, cfg.QuerySleepTime)
			},
		}, {
			name: "Test_Config_Max_Txn_Query_Less_Than_1",

			setup: func(t *testing.T) Reader {

				return mockDefaultReader()
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(5, cfg.MaxTxnQuery)
			},
		}, {
			name: "Test_Config_Confirmation_Chain_Length_Less_Than_1",

			setup: func(t *testing.T) Reader {
				return mockDefaultReader()
			},
			run: func(r *require.Assertions, cfg Config) {
				r.Equal(3, cfg.ConfirmationChainLength)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			require := require.New(t)

			reader := tt.setup(t)

			cfg, err := LoadConfig(reader)

			// test it by predefined error variable instead of error message
			if tt.exceptedErr != nil {
				require.ErrorIs(err, tt.exceptedErr)
			} else {
				require.Equal(nil, err)
			}

			if tt.run != nil {
				tt.run(require, cfg)
			}

		})
	}
}
