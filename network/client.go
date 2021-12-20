package net

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	cfg "github.com/vecno-io/go-pyteal/config"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/types"
)

func MakeClient() (*algod.Client, error) {
	path := cfg.DataPath()
	// Fix hard coded sub path
	if cfg.Devnet == cfg.Target() {
		path += "/primary"
	}

	addr, err := getFirstLineFromFile(fmt.Sprintf(
		"%s/algod.net", path,
	))
	if err != nil {
		return nil, fmt.Errorf("read network file: %s", err)
	}

	token, err := getFirstLineFromFile(fmt.Sprintf(
		"%s/algod.token", path,
	))
	if err != nil {
		return nil, fmt.Errorf("read token file: %s", err)
	}
	return algod.MakeClient("http://"+addr, token)
}

func MakeTxnParams() (types.SuggestedParams, error) {
	cln, err := MakeClient()
	if err != nil {
		return types.SuggestedParams{}, fmt.Errorf("make client: %s", err)
	}

	txnParams, err := cln.SuggestedParams().Do(context.Background())
	if err != nil {
		return types.SuggestedParams{}, fmt.Errorf("suggested params: %s", err)
	}
	return txnParams, nil
}

func SendRawTransaction(txn []byte) (txInfo models.PendingTransactionInfoResponse, err error) {
	cln, err := MakeClient()
	if err != nil {
		err = fmt.Errorf("make client: %s", err)
		return
	}

	pendingTxID, err := cln.SendRawTransaction(txn).Do(context.Background())
	if err != nil {
		err = fmt.Errorf("client send: %s", err)
		return
	}

	txInfo, err = WaitForConfirmation(cln, pendingTxID, 24, context.Background())
	if err != nil {
		err = fmt.Errorf("client wait: %s", err)
		return
	}
	if len(txInfo.PoolError) > 0 {
		err = fmt.Errorf("txn pool: %s", err)
		return
	}
	return
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
	return lines[0], nil
}
