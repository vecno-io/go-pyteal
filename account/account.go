package acc

import (
	"fmt"
	"os"

	cfg "github.com/vecno-io/go-pyteal/config"

	"github.com/algorand/go-algorand-sdk/crypto"
)

func Load(name, pass string) (crypto.Account, error) {
	fmt.Println(":: Load account:", cfg.DataPath())

	path := fmt.Sprintf("%s/%s.acc", cfg.DataPath(), name)

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
	fmt.Println(":: Create account:", cfg.DataPath())

	path := fmt.Sprintf("%s/%s.acc", cfg.DataPath(), name)

	if doesAccountExist(path) {
		return crypto.Account{}, fmt.Errorf("create account: file exists: %s", path)
	}

	acc := crypto.GenerateAccount()
	if err := SaveAccountToFile(acc, pass, path); nil != err {
		return crypto.Account{}, fmt.Errorf("create account: save file : %s", err)
	}

	return acc, nil
}

func doesAccountExist(file string) bool {
	if _, err := os.Stat(file); nil == err {
		return true
	}
	return false
}