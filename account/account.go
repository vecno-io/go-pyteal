package acc

import (
	"fmt"

	cfg "github.com/vecno-io/go-pyteal/config"
)

func Create(name, pass string) error {
	fmt.Println(":: Create account:", cfg.DataPath())

	return nil
}
