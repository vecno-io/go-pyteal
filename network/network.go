package net

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	cfg "github.com/vecno-io/go-pyteal/config"
)

func Start() error {
	fmt.Println(":: Start network:", cfg.DataPath())

	if cfg.Testnet == cfg.Target() || cfg.Mainnet == cfg.Target() {
		return startNetworkPub()
	}
	return startNetworkPriv()
}

func Stop() error {
	fmt.Println(":: Stop network:", cfg.DataPath())

	var cmd string
	if cfg.Testnet == cfg.Target() || cfg.Mainnet == cfg.Target() {
		cmd = fmt.Sprintf("goal node stop -d %s", cfg.DataPath())
	} else {
		cmd = fmt.Sprintf("goal network stop -r %s", cfg.DataPath())
	}

	fmt.Println(">>", cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if nil != err {
		return err
	}

	return nil
}

func Status() error {
	fmt.Println(":: Status network:", cfg.DataPath())

	var cmd string
	if cfg.Testnet == cfg.Target() || cfg.Mainnet == cfg.Target() {
		cmd = fmt.Sprintf("goal node status -d %s", cfg.DataPath())
	} else {
		cmd = fmt.Sprintf("goal network status -r %s", cfg.DataPath())
	}

	fmt.Println(">>", cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if nil != err {
		return err
	}

	return nil
}

func Create() error {
	if _, err := os.Stat(cfg.DataPath()); nil == err {
		return fmt.Errorf("create network: path already exists")
	}
	fmt.Println(":: Create network:", cfg.DataPath())

	if cfg.Testnet == cfg.Target() {
		return createNetworkPub("genesisfiles/testnet/genesis.json")
	} else if cfg.Mainnet == cfg.Target() {
		return createNetworkPub("genesisfiles/mainnet/genesis.json")
	}
	return createNetworkPriv()
}

func Destroy() error {
	fmt.Println(":: Destroy network:", cfg.DataPath())

	if cfg.Testnet == cfg.Target() || cfg.Mainnet == cfg.Target() {
		return destroyNetworkPub()
	}
	return destroyNetworkPriv()
}

func IsActive() bool {
	if _, err := os.Stat(fmt.Sprintf(
		"%s/algod.pid", cfg.DataPath(),
	)); err != nil {
		return false
	}
	return true
}

func startNetworkPub() error {
	var url string
	switch cfg.Target() {
	case cfg.Testnet:
		url = "https://algorand-catchpoints.s3.us-east-2.amazonaws.com/channel/testnet/latest.catchpoint"
	case cfg.Mainnet:
		url = "https://algorand-catchpoints.s3.us-east-2.amazonaws.com/channel/mainnet/latest.catchpoint"
	default:
		return fmt.Errorf("unsuported network type")
	}
	point, err := loadCatchPoint(url)
	if nil != err {
		return err
	}

	cmd := fmt.Sprintf("goal node start -d %s", cfg.DataPath())
	fmt.Println(">>", cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if nil != err {
		return err
	}

	// Hack, node needs to load
	fmt.Println(":: Loading catchup target")
	time.Sleep(5 * time.Second)
	// Hack, before it can catchup

	cmd = fmt.Sprintf("goal node -d %s catchup %s", cfg.DataPath(), point)
	fmt.Println(">>", cmd)
	out, err = exec.Command("bash", "-c", cmd).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if nil != err {
		return err
	}

	return nil
}

func startNetworkPriv() error {
	cmd := fmt.Sprintf("goal network start -r %s", cfg.DataPath())
	fmt.Println(">>", cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if nil != err {
		return err
	}
	return nil
}

func createNetworkPub(srcPath string) error {
	if err := os.Mkdir(cfg.DataPath(), 0755); err != nil {
		return fmt.Errorf("create network: failed to make path %s", err)
	}
	source, err := os.Open(fmt.Sprintf("%s/%s", cfg.NodePath(), srcPath))
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(fmt.Sprintf("%s/genesis.json", cfg.DataPath()))
	if err != nil {
		return err
	}
	defer destination.Close()
	if _, err := io.Copy(destination, source); nil != err {
		return fmt.Errorf("create network: failed to copy file: %s", err)
	}

	// Enable the developers api to compile teal code
	cfgFile := fmt.Sprintf("%s/config.json", cfg.DataPath())
	if err := ioutil.WriteFile(
		cfgFile, []byte(`{"EnableDeveloperAPI":true}`), os.ModePerm,
	); nil != err {
		return fmt.Errorf("create network: failed write config: %s", err)
	}
	return nil
}

func createNetworkPriv() error {
	cfgFile := fmt.Sprintf("%s/network.json", cfg.NodePath())
	if _, err := loadPrivateNetworkConfig(cfgFile); nil != err {
		return fmt.Errorf("create networ: load config: %s", err)
	}
	cmd := fmt.Sprintf(
		"goal network create -n devnet -t %s -r %s",
		cfgFile, cfg.DataPath(),
	)
	fmt.Println(">>", cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if nil != err {
		return err
	}

	node := cfg.NodeConfig{}
	// Enable the developers api to compile teal code
	// ToDo fix the hard coded node path below, none default
	cfgFile = fmt.Sprintf("%s/primary/config.json", cfg.DataPath())
	file, err := os.ReadFile(cfgFile)
	if err := json.Unmarshal(file, &node); nil != err {
		return err
	}
	node.EnableDeveloperAPI = true
	jsonString, _ := json.Marshal(node)
	if os.WriteFile(cfgFile, jsonString, os.ModePerm); nil != err {
		return err
	}

	return nil
}

func destroyNetworkPub() error {
	fmt.Println(">>", fmt.Sprintf("goal network delete -r %s", cfg.DataPath()))
	cmd := fmt.Sprintf("goal network stop -r %s", cfg.DataPath())
	exec.Command("bash", "-c", cmd).Output()
	return os.RemoveAll(cfg.DataPath())
}

func destroyNetworkPriv() error {
	cmd := fmt.Sprintf("goal network delete -r %s", cfg.DataPath())
	fmt.Println(">>", cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if nil != err {
		return err
	}
	return nil
}

func loadCatchPoint(path string) (string, error) {
	resp, err := http.Get(path)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Can this be validated somehow?
	return strings.TrimSpace(string(body)), nil
}

func loadPrivateNetworkConfig(filePath string) ([]byte, error) {
	if data, err := os.ReadFile(filePath); err == nil {
		return data, nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()

	defaultNetwork := cfg.DefaultNetwork()
	if _, err := file.Write(defaultNetwork); err != nil {
		return []byte{}, err
	}
	file.Sync()

	return defaultNetwork, nil
}
