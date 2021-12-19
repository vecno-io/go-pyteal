package contract

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	cfg "github.com/vecno-io/go-pyteal/config"
	net "github.com/vecno-io/go-pyteal/network"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
)

type Setup struct {
	Manager crypto.Account

	ClearProg    string
	ApprovalProg string

	LocalSchema  types.StateSchema
	GlobalSchema types.StateSchema
}

func GetId(name string) (uint64, error) {
	return loadFromJsonFile(name)
}

func Deploy(s Setup) error {
	optIn := true

	if _, err := os.Stat(fmt.Sprintf(
		"%s/%s.id",
		cfg.AssetPath(), s.ApprovalProg),
	); nil == err {
		return fmt.Errorf("deploy: %s is already deployed", s.ApprovalProg)
	}
	fmt.Println(":: Deploy contract build:", s.ApprovalProg)

	clearProg, err := ioutil.ReadFile(fmt.Sprintf(
		"%s/contracts/%s.prog", cfg.AssetPath(), s.ClearProg,
	))
	if err != nil {
		return fmt.Errorf("deploy failed: %s: read file: %s", s.ClearProg, err)
	}
	approvalProg, err := ioutil.ReadFile(fmt.Sprintf(
		"%s/contracts/%s.prog", cfg.AssetPath(), s.ApprovalProg,
	))
	if err != nil {
		return fmt.Errorf("deploy failed: %s: read file: %s", s.ApprovalProg, err)
	}

	appArgs := [][]byte{}
	accounts := []string{}
	foreignApps := []uint64{}
	foreignAssets := []uint64{}

	cln, err := net.MakeClient()
	if err != nil {
		return fmt.Errorf("deploy failed: make client: %s", err)
	}
	txnParams, err := cln.SuggestedParams().Do(context.Background())
	if err != nil {
		return fmt.Errorf("deploy failed: suggested params: %s", err)
	}

	note := []byte{}
	group := types.Digest{}
	lease := [32]byte{}
	rekeyTo := types.ZeroAddress
	extraPages := uint32(0)

	createTx, err := future.MakeApplicationCreateTxWithExtraPages(
		optIn, approvalProg, clearProg, s.GlobalSchema, s.LocalSchema,
		appArgs, accounts, foreignApps, foreignAssets, txnParams,
		s.Manager.Address, note, group, lease, rekeyTo, extraPages,
	)
	if err != nil {
		return fmt.Errorf("deploy failed: make create tx: %s", err)
	}

	// Enforce it or fail, a bug?
	createTx.OnCompletion = types.OptInOC

	_, signedTx, err := crypto.SignTransaction(s.Manager.PrivateKey, createTx)
	if err != nil {
		return fmt.Errorf("deploy failed: sign create tx: %s", err)
	}
	pendingTx, err := cln.SendRawTransaction(signedTx).Do(context.Background())
	if err != nil {
		return fmt.Errorf("deploy failed: send create tx: %s", err)
	}

	txConfirm, err := net.WaitForConfirmation(cln, pendingTx, 24, context.Background())
	if err != nil {
		return fmt.Errorf("deploy failed: confirm tx: %s", err)
	}
	if len(txConfirm.PoolError) > 0 {
		return fmt.Errorf("deploy failed: pool error: %s", txConfirm.PoolError)
	}

	fmt.Printf(">> App deployed with id: %d\n", txConfirm.ApplicationIndex)
	if err := saveToFile(s.ApprovalProg, txConfirm.ApplicationIndex); err != nil {
		return fmt.Errorf("contract: failed to save app: %s", err)
	}

	return nil
}

func saveToFile(name string, id uint64) error {
	str, err := json.Marshal(id)
	if err != nil {
		return err
	}
	// TODO Fix Path
	if err = ioutil.WriteFile(fmt.Sprintf(
		"%s/%s.id", cfg.AssetPath(), name,
	), str, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func loadFromJsonFile(name string) (uint64, error) {
	id := uint64(0)
	// TODO Fix Path
	data, err := os.ReadFile(fmt.Sprintf(
		"%s/%s.id", cfg.AssetPath(), name,
	))
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal(data[:], &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
