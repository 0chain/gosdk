package conf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func setUpConfig(name, content string) {

	ioutil.WriteFile(filepath.Join(getConfigDir(), name), []byte(content), 0744)
}

func tearDownConfig(name string) {
	os.Remove(filepath.Join(getConfigDir(), name))
}

func TestMissingConfig(t *testing.T) {

	configFile := "missing_config.yaml"

	err := Load(configFile)

	require.ErrorIs(t, err, ErrMssingConfig)
}

func TestBadFormat(t *testing.T) {

	setUpConfig("bad_format.yaml", `
{
	block_worker:"",

`)
	defer tearDownConfig("bad_format.yaml")

	err := Load("bad_format.yaml")

	require.ErrorIs(t, err, ErrBadFormat)
}

func TestInvalidBlockWorker(t *testing.T) {

	setUpConfig("invalid_blockworker.yaml", `
block_worker: 127.0.0.1:9091
`)
	defer tearDownConfig("invalid_blockworker.yaml")

	err := Load("invalid_blockworker.yaml")

	require.ErrorIs(t, err, ErrInvalidValue)
}
