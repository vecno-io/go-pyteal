package contract

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/algorand/go-algorand-sdk/logic"

	cfg "github.com/vecno-io/go-pyteal/config"
	net "github.com/vecno-io/go-pyteal/network"
)

func Build(list []string) error {
	fmt.Println(":: Contracts build:", cfg.DataPath())

	for _, s := range list {
		if err := build(s); nil != err {
			return err
		}
		if err := compile(s); nil != err {
			return err
		}
	}
	return nil
}

func build(name string) error {
	path := fmt.Sprintf("%s/contracts/%s", cfg.AssetPath(), name)
	cmd := fmt.Sprintf("python3 %s.py > %s.teal", path, path)
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

func compile(name string) error {
	cln, err := net.MakeClient()
	if err != nil {
		return fmt.Errorf("compile %s failed: make client: %s", name, err)
	}

	path := fmt.Sprintf("%s/contracts/%s", cfg.AssetPath(), name)
	fmt.Println(">> compile", path)

	teal, err := ioutil.ReadFile(fmt.Sprintf("%s.teal", path))
	if err != nil {
		return fmt.Errorf("compile %s failed: read file: %s", name, err)
	}

	chk, err := cln.TealCompile(teal).Do(context.Background())
	if err != nil {
		return fmt.Errorf("compile %s failed: compile teal: %s", name, err)
	}
	prg, err := base64.StdEncoding.DecodeString(chk.Result)
	if err != nil {
		return fmt.Errorf("compile %s failed: decode program: %s", name, err)
	}
	err = logic.CheckProgram(prg, make([][]byte, 0))
	if nil != err {
		return fmt.Errorf("compile %s failed: check program: %s", name, err)
	}

	if err = ioutil.WriteFile(fmt.Sprintf("%s.prog", path), prg, os.ModePerm); nil != err {
		return fmt.Errorf("compile %s failed: write file: %s", name, err)
	}

	return nil
}
