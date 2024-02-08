package conf

import (
	"errors"
	"os"
	"strings"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/sys"
	"github.com/spf13/viper"
)

// Network settings from ~/.zcn/network.yaml
type Network struct {
	// Sharders sharder list of blockchain
	sharders []string
	// Miners miner list of blockchain
	miners []string
}

func NewNetwork(miners, sharders []string) (*Network, error) {
	n := &Network{
		miners: miners,
		sharders: sharders,
	}
	if !n.IsValid() {
		return nil, errors.New("network has no miners/sharders")
	}
	n.NormalizeURLs()
	return n, nil
}

// IsValid check network if it has miners and sharders
func (n *Network) IsValid() bool {
	return n != nil && len(n.miners) > 0 && len(n.sharders) > 0
}

func (n *Network) Miners() []string {
	return n.miners
}

func (n *Network) Sharders() []string {
	return n.sharders
}

func (n *Network) NormalizeURLs() {
	for i := 0; i < len(n.miners); i++ {
		n.miners[i] = strings.TrimSuffix(n.miners[i], "/")
	}

	for i := 0; i < len(n.sharders); i++ {
		n.sharders[i] = strings.TrimSuffix(n.sharders[i], "/")
	}
}

// LoadNetworkFile load and parse Network from file
func LoadNetworkFile(file string) (Network, error) {

	var network Network
	var err error

	_, err = sys.Files.Stat(file)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return network, thrown.Throw(ErrMssingConfig, file)
		}
		return network, err
	}

	v := viper.New()

	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return network, thrown.Throw(ErrBadParsing, err.Error())
	}

	return LoadNetwork(v), nil
}

// LoadNetwork load and parse network
func LoadNetwork(v Reader) Network {
	return Network{
		sharders: v.GetStringSlice("sharders"),
		miners:   v.GetStringSlice("miners"),
	}
}
