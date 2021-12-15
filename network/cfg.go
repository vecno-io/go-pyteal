package net

import (
	"fmt"
	"os"
)

type NetType int

type Config struct {
	Type     NetType `mapstructure:"ALGORAND_TYPE"`
	NodePath string  `mapstructure:"ALGORAND_NODE"`
	DataPath string  `mapstructure:"ALGORAND_DATA"`
}

type NodeConfig struct {
	Version            uint64
	GossipFanout       uint64
	NetAddress         string
	DNSBootstrapID     string
	EnableProfiler     bool
	EnableDeveloperAPI bool
}

const (
	Devnet  NetType = 1 << iota // == 1
	Testnet NetType = 1 << iota // == 2
	Mainnet NetType = 1 << iota // == 3
)

var cfg = Config{
	Type:     Devnet,
	NodePath: "~/node",
	DataPath: "~/node/devnet-data",
}

func setConfig(c Config) error {
	if c.Type == 0 || c.Type > 3 {
		return fmt.Errorf("unknown type: NetType: %d", c.Type)
	}
	cfg = c
	return nil
}

// isNodePath fast sanity check to avoid issues later on
func isNodePath(path string) error {
	info, err := os.Stat(path)
	if nil != err {
		return fmt.Errorf("invalid path: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: %s", "not a directory")
	}

	// Check for required tool set
	info, err = os.Stat(fmt.Sprintf("%s/kdm", path))
	if nil != err {
		return fmt.Errorf("invalid path: kdm: %s", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid path: kdm: %s", "not a file")
	}
	info, err = os.Stat(fmt.Sprintf("%s/goal", path))
	if nil != err {
		return fmt.Errorf("invalid path: goal: %s", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid path: goal: %s", "not a file")
	}
	info, err = os.Stat(fmt.Sprintf("%s/algod", path))
	if nil != err {
		return fmt.Errorf("invalid path: algod: %s", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid path: algod: %s", "not a file")
	}

	// Check for required configuration file if needed
	switch cfg.Type {
	case Devnet:
		return nil
	case Testnet:
		info, err = os.Stat(fmt.Sprintf("%s/genesisfiles/testnet/genesis.json", path))
	case Mainnet:
		info, err = os.Stat(fmt.Sprintf("%s/genesisfiles/mainnet/genesis.json", path))
	default:
		return fmt.Errorf("invalid path: unknown network type")
	}

	if nil != err {
		return fmt.Errorf("invalid path: config: %s", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid path: config: %s", "not a file")
	}

	return nil
}

// isNetworkPath fast sanity check to avoid issues later on
func isNetworkPath(path string) error {
	info, err := os.Stat(path)
	if nil != err {
		return fmt.Errorf("invalid path: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: %s", "not a directory")
	}
	info, err = os.Stat(fmt.Sprintf("%s/genesis.json", path))
	if nil != err {
		return fmt.Errorf("invalid path: genesis: %s", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid path: genesis: %s", "not a file")
	}
	return nil
}

var defaultNetworkConfig = []byte(`{
    "Genesis": {
        "NetworkName": "private",
        "ConsensusProtocol": "future",
        "Wallets": [
            {
                "Name": "wallet",
                "Stake": 100,
                "Online": true
            }
        ]
    },
    "Nodes": [
        {
            "Name": "primary",
            "IsRelay": true,
            "Wallets": [
                {
                    "Name": "wallet",
                    "ParticipationOnly": false
                }
            ]
        }
    ]
}`)
