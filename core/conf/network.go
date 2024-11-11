package conf

import (
	"errors"
	"os"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/sys"
	"github.com/spf13/viper"
)

// Network settings from ~/.zcn/network.yaml
type Network struct {
	// Sharders sharder list of blockchain
	Sharders []string
	// Miners miner list of blockchain
	Miners []string
}

// IsValid check network if it has miners and sharders
func (n *Network) IsValid() bool {
	return n != nil && len(n.Miners) > 0 && len(n.Sharders) > 0
}

// LoadNetworkFile load and parse Network from file
//   - file is the path of the file (full path)
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
		Sharders: v.GetStringSlice("sharders"),
		Miners:   v.GetStringSlice("miners"),
	}
}
