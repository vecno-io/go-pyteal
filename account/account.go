package acc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	cfg "github.com/vecno-io/go-pyteal/config"
	net "github.com/vecno-io/go-pyteal/network"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
)

func Info(name, pass string) (models.Account, error) {
	path := fmt.Sprintf("%s/accounts/%s.acc", cfg.AssetPath(), name)
	fmt.Println(":: Load account:", path)

	if !doesAccountExist(path) {
		return models.Account{}, fmt.Errorf("account info: not found: %s", path)
	}

	cl, err := net.MakeClient()
	if err != nil {
		return models.Account{}, fmt.Errorf("account info: make client: %s", err)
	}

	ac, err := LoadAccountFromFile(pass, path)
	if nil != err {
		return models.Account{}, fmt.Errorf("account info: load: %s", err)
	}

	info, err := cl.AccountInformation(ac.Address.String()).Do(context.Background())
	if err != nil {
		return models.Account{}, fmt.Errorf("account info: get: %s", err)
	}

	return info, nil
}

func Load(name, pass string) (crypto.Account, error) {
	path := fmt.Sprintf("%s/accounts/%s.acc", cfg.AssetPath(), name)
	fmt.Println(":: Load account:", path)

	if !doesAccountExist(path) {
		return crypto.Account{}, fmt.Errorf("load account: account not found: %s", path)
	}

	acc, err := LoadAccountFromFile(pass, path)
	if nil != err {
		return crypto.Account{}, fmt.Errorf("load account: %s", err)
	}

	return acc, nil
}

func Create(name, pass string) (crypto.Account, error) {
	path := fmt.Sprintf("%s/accounts/%s.acc", cfg.AssetPath(), name)
	fmt.Println(":: Create account:", path)

	if doesAccountExist(path) {
		return crypto.Account{}, fmt.Errorf("create account: file exists: %s", path)
	}

	acc := crypto.GenerateAccount()
	if err := SaveAccountToFile(acc, pass, path); nil != err {
		return crypto.Account{}, fmt.Errorf("create account: save file : %s", err)
	}

	return acc, nil
}

func DevFunding(address string, amount uint64) error {
	if cfg.Target() != cfg.Devnet {
		return fmt.Errorf("funding: only available for devnet")
	}
	seed, err := getDevSeedAddress()
	if nil != err {
		return fmt.Errorf("funding: get seed: %s", err)
	}

	cmd := fmt.Sprintf(
		"goal -d %s/%s clerk send -a %d -f %s -t %s",
		cfg.DataPath(), "primary", amount, seed, address,
	)
	fmt.Println(">>", cmd)
	exec.Command("bash", "-c", cmd).Output()

	return nil
}

func doesAccountExist(file string) bool {
	if _, err := os.Stat(file); nil == err {
		return true
	}
	return false
}

func getDevSeedAddress() (string, error) {
	cmd := fmt.Sprintf("goal account list -d %s/%s | awk '{ print $3 }' | head -n 1", cfg.DataPath(), "primary")
	fmt.Println(">>", cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	return strings.TrimSpace(string(out)), err
}
