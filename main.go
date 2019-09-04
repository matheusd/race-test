package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/decred/dcrd/chaincfg/v2"
	"github.com/decred/dcrd/dcrutil/v2"
	"github.com/decred/dcrd/rpctest"
	"github.com/decred/dcrd/wire"

	"github.com/decred/dcrwallet/chain/v3"
	walletloader "github.com/decred/dcrwallet/loader"
	base "github.com/decred/dcrwallet/wallet/v3"
	"github.com/decred/dcrwallet/wallet/v3/txrules"
)

var (
	netParams    = chaincfg.RegNetParams()
	nullArray    = [128]byte{}
	walletSynced bool
)

func fatalf(msg string, args ...interface{}) {
	panic(fmt.Errorf(msg, args...))
}

func logf(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	fmt.Printf("\n")
}

func onRpcSyncerSynced(synced bool) {
	logf("RPC Syncer synced")
	walletSynced = true
}

func walletBalance(wallet *base.Wallet) dcrutil.Amount {
	balances, err := wallet.CalculateAccountBalance(0, 1)
	if err != nil {
		fatalf("error getting balance: %v", err)
	}
	return balances.Spendable
}

func newWalletAddressScript(wallet *base.Wallet) []byte {
	addr, err := wallet.NewExternalAddress(context.Background(), 0)
	if err != nil {
		panic(err)
	}
	scripter := addr.(base.V0Scripter)
	return scripter.ScriptV0()
}

func tryFundingWallet(wallet *base.Wallet, miningNode *rpctest.Harness) {
	amount := dcrutil.Amount(1e5)

	initialBalance := walletBalance(wallet)

	for i := 0; i < 60; i++ {
		script := newWalletAddressScript(wallet)

		output := &wire.TxOut{
			Value:    int64(amount),
			PkScript: script,
			Version:  0,
		}
		_, err := miningNode.SendOutputs([]*wire.TxOut{output}, 1e5)
		if err != nil {
			panic(err)
		}
		logf("Sent output %d", i)
	}

	_ = initialBalance
}

func main() {
	// Create the rpctest harness main node.
	miningNode, err := rpctest.New(netParams, nil, []string{"--txindex"})
	if err != nil {
		fatalf("unable to create mining node: %v", err)
	}
	logf("mining node created")

	if err := miningNode.SetUp(true, 60); err != nil {
		fatalf("unable to set up mining node: %v", err)
	}
	logf("mining node setup")
	defer miningNode.TearDown()

	// Create the new test wallet
	tempTestDir, err := ioutil.TempDir("", "test-wallet")
	if err != nil {
		fatalf("unable to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempTestDir)
	loader := walletloader.NewLoader(netParams, tempTestDir,
		&walletloader.StakeOptions{}, base.DefaultGapLimit, false,
		txrules.DefaultRelayFeePerKb.ToCoin(), base.DefaultAccountGapLimit,
		false)

	wallet, err := loader.CreateNewWallet([]byte("public"), []byte("private"),
		nullArray[:32])
	if err != nil {
		panic(err)
	}
	logf("wallet created")

	if err := wallet.Unlock([]byte("private"), nil); err != nil {
		panic(err)
	}
	logf("wallet unlocked")

	// Create the chain.RPCClient that we'll use to connect to the wallet
	rpcConfig := miningNode.RPCConfig()
	chainRpcOpts := chain.RPCOptions{
		Address: rpcConfig.Host,
		User:    rpcConfig.User,
		Pass:    rpcConfig.Pass,
		CA:      rpcConfig.Certificates,
	}
	walletChainSyncer := chain.NewSyncer(wallet, &chainRpcOpts)
	walletChainSyncer.SetCallbacks(&chain.Callbacks{
		Synced: onRpcSyncerSynced,
	})
	logf("wallet Chain RPC Syncer created")

	go func() {
		logf("Starting syncer...")
		ctx := context.TODO()
		err := walletChainSyncer.Run(ctx)

		if err != nil {
			logf("error after syncer.run: %v", err)
		}
	}()

	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		if walletSynced {
			break
		}
	}
	ticker.Stop()

	tryFundingWallet(wallet, miningNode)

	logf("Done!")
}
