package main

import (
	"context"
	"fmt"
	"os"
	"io/ioutil"
	"time"

	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrd/rpctest"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrd/txscript"

	"github.com/decred/dcrwallet/chain"
	walletloader "github.com/decred/dcrwallet/loader"
	base "github.com/decred/dcrwallet/wallet"
	"github.com/decred/dcrwallet/wallet/txrules"
	"github.com/decred/dcrwallet/errors"
)

var (
	netParams = &chaincfg.RegNetParams
	nullArray = [128]byte{}
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

func newWalletAddress(wallet *base.Wallet) dcrutil.Address {
	addr, err := wallet.NewExternalAddress(0)
	if err != nil {
		panic(err)
	}
	return addr
}

func tryFundingWallet(wallet *base.Wallet, miningNode *rpctest.Harness) {
	amount := dcrutil.Amount(1e6)

	initialBalance := walletBalance(wallet)

	for i := 0; i < 20; i++ {
		addr := newWalletAddress(wallet)
		script, err := txscript.PayToAddrScript(addr)
		if err != nil {
			panic(err)
		}

		output := &wire.TxOut{
			Value:    int64(amount),
			PkScript: script,
			Version: txscript.DefaultScriptVersion,
		}
		_, err = miningNode.SendOutputs([]*wire.TxOut{output}, 1e5);
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

	if err := miningNode.SetUp(true, 25); err != nil {
		fatalf("unable to set up mining node: %v", err)
	}
	logf("mining node setup")
	defer miningNode.TearDown()

	// Create the chain.RPCClient that we'll use to connect to the wallet
	rpcConfig := miningNode.RPCConfig()
	walletRpcClient, err := chain.NewRPCClient(netParams,
		rpcConfig.Host, rpcConfig.User, rpcConfig.Pass,
		rpcConfig.Certificates, false)
	if err != nil {
		fatalf("unable to make chain rpc: %v", err)
	}
	logf("wallet RPCClient created")

	// Create the new test wallet
	tempTestDir, err := ioutil.TempDir("", "test-wallet")
	if err != nil {
		fatalf("unable to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempTestDir)
	loader := walletloader.NewLoader(netParams, tempTestDir,
		&walletloader.StakeOptions{}, base.DefaultGapLimit, false,
		txrules.DefaultRelayFeePerKb.ToCoin(), base.DefaultAccountGapLimit)

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

	err = walletRpcClient.Start(context.TODO(), true)
	if err != nil && !errors.MatchAll(rpcclient.ErrClientAlreadyConnected, err) {
		panic(err)
	}
	logf("wallet rpcclient started")

	walletNetBackend := chain.BackendFromRPCClient(walletRpcClient.Client)
	wallet.SetNetworkBackend(walletNetBackend)

	go func () {
		logf("Starting syncer...")
		syncer := chain.NewRPCSyncer(wallet, walletRpcClient)
		syncer.SetNotifications(&chain.Notifications{
			Synced: onRpcSyncerSynced,
		})
		ctx := context.TODO()
		err := syncer.Run(ctx, true)

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