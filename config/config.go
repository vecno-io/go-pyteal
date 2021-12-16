package cfg

import (
	"fmt"
	"os"
)

type Network int

const (
	Devnet  Network = 1 << iota
	Testnet Network = 1 << iota
	Mainnet Network = 1 << iota
)

type Setup struct {
	Target     string `mapstructure:"type"`
	Timeout    uint32 `mapstructure:"time"`
	NodePath   string `mapstructure:"node"`
	Passphrase string `mapstructure:"pass"`
}

type Config struct {
	Target   Network
	Timeout  uint32
	NodePath string
	DataPath string
}

var cfg = Config{
	Timeout: 16,
}

func Target() Network {
	return cfg.Target
}

func Timeout() uint32 {
	return cfg.Timeout
}

func NodePath() string {
	return cfg.NodePath
}

func DataPath() string {
	return cfg.DataPath
}

func OnCreate(s Setup) error {
	if s.Timeout > 16 {
		cfg.Timeout = s.Timeout
	}

	switch s.Target {
	case "devnet":
		cfg.Target = Devnet
	case "testnet":
		cfg.Target = Testnet
	case "mainnet":
		cfg.Target = Mainnet
	default:
		return fmt.Errorf("init config: unknown type: %s", s.Target)
	}

	return nil
}

func OnInitialize(s Setup, validate bool) error {
	if s.Timeout > 16 {
		cfg.Timeout = s.Timeout
	}

	switch s.Target {
	case "devnet":
		cfg.Target = Devnet
	case "testnet":
		cfg.Target = Testnet
	case "mainnet":
		cfg.Target = Mainnet
	default:
		return fmt.Errorf("init config: unknown type: %s", s.Target)
	}

	if err := IsNodePath(s.NodePath, cfg.Target); nil == err {
		cfg.NodePath = s.NodePath
	} else if path, err := os.UserHomeDir(); nil == err {
		cfg.NodePath = fmt.Sprintf("%s/node", path)
	}
	if err := IsNodePath(cfg.NodePath, cfg.Target); nil != err {
		return fmt.Errorf("init config: unable to find node: %s", s.NodePath)
	}

	switch cfg.Target {
	case Devnet:
		cfg.DataPath = fmt.Sprintf("%s/devnet-data", cfg.NodePath)
	case Testnet:
		cfg.DataPath = fmt.Sprintf("%s/testnet-data", cfg.NodePath)
	case Mainnet:
		cfg.DataPath = fmt.Sprintf("%s/mainnet-data", cfg.NodePath)
	default:
		return fmt.Errorf("init config: unknown traget: %d", cfg.Target)
	}

	if err := IsNetworkPath(cfg.DataPath, cfg.Target); validate && nil != err {
		return fmt.Errorf("init config: invalid network path: %s", err)
	}

	return nil
}

func IsNodePath(path string, target Network) error {
	info, err := os.Stat(path)
	if nil != err {
		return fmt.Errorf("invalid path: %s", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid path: not a directory: %s", path)
	}

	// Check for required tool set
	info, err = os.Stat(fmt.Sprintf("%s/kmd", path))
	if nil != err {
		return fmt.Errorf("invalid path: kmd: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: kmd: not a file: %s/algod", path)
	}
	info, err = os.Stat(fmt.Sprintf("%s/goal", path))
	if nil != err {
		return fmt.Errorf("invalid path: goal: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: goal: not a file: %s/algod", path)
	}
	info, err = os.Stat(fmt.Sprintf("%s/algod", path))
	if nil != err {
		return fmt.Errorf("invalid path: algod: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: algod: not a file: %s/algod", path)
	}
	var file string
	switch target {
	case Devnet:
		return nil
	case Testnet:
		file = fmt.Sprintf("%s/genesisfiles/testnet/genesis.json", path)
	case Mainnet:
		file = fmt.Sprintf("%s/genesisfiles/mainnet/genesis.json", path)
	default:
		return fmt.Errorf("invalid path: unknown target")
	}

	info, err = os.Stat(file)
	if nil != err {
		return fmt.Errorf("invalid path: genesis: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: genesis: not a file: %s", file)
	}

	return nil
}

func IsNetworkPath(path string, target Network) error {
	info, err := os.Stat(path)
	if nil != err {
		return fmt.Errorf("invalid path: %s", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid path: not a directory: %s", path)
	}

	if Devnet == target {
		// Fix hard coded
		path += "/primary"
	}
	info, err = os.Stat(fmt.Sprintf("%s/config.json", path))
	if nil != err {
		return fmt.Errorf("invalid path: config: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: config: not a file: %s/config.json", path)
	}
	info, err = os.Stat(fmt.Sprintf("%s/genesis.json", path))
	if nil != err {
		return fmt.Errorf("invalid path: genesis: %s", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid path: genesis: not a file: %s/genesis.json", path)
	}
	return nil
}

func DefaultNetwork() []byte {
	return []byte(`{
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
}
