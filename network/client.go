package net

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

func MakeClient(c Config) (*algod.Client, error) {
	if err := setConfig(c); nil != err {
		return nil, fmt.Errorf("make client: set: %s", err)
	}
	if err := verifyNetworkPath(); nil == err {
		return nil, fmt.Errorf("make client: verify: %s", err)
	}

	fmt.Println(":: Make client:", cfg.DataPath)

	path := cfg.DataPath
	if Devnet == cfg.Type {
		path += "/primary"
	}

	addr, err := getFirstLineFromFile(fmt.Sprintf(
		"%s/algod.net", path,
	))
	if err != nil {
		return nil, fmt.Errorf("read network file: %v", err)
	}

	token, err := getFirstLineFromFile(fmt.Sprintf(
		"%s/algod.token", path,
	))
	if err != nil {
		return nil, fmt.Errorf("read token file: %v", err)
	}
	return algod.MakeClient("http://"+addr, token)
}

func WaitForConfirmation(c *algod.Client, txid string, waitRounds uint64, ctx context.Context, headers ...*common.Header) (txInfo models.PendingTransactionInfoResponse, err error) {
	response, err := c.Status().Do(ctx, headers...)
	if err != nil {
		return
	}

	lastRound := response.LastRound
	currentRound := lastRound + 1

	for {
		// Check that the `waitRounds` has not passed
		if waitRounds > 0 && currentRound > lastRound+waitRounds {
			err = fmt.Errorf("wait for transaction id %s timed out", txid)
			return
		}
		txInfo, _, err = c.PendingTransactionInformation(txid).Do(ctx, headers...)
		if err != nil {
			return
		}
		// The transaction has been confirmed
		if txInfo.ConfirmedRound > 0 {
			return
		}
		// Wait until the block for the `currentRound` is confirmed
		response, err = c.StatusAfterBlock(currentRound).Do(ctx, headers...)
		if err != nil {
			return
		}
		// Increment the `currentRound`
		currentRound += 1
	}
}

func getFirstLineFromFile(file string) (string, error) {
	addrStr, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(addrStr), "\n")
	return lines[0], err
}