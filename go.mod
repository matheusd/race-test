module main

go 1.12

require (
	github.com/decred/dcrd v1.2.1-0.20190817061140-8497b9843bcb
	github.com/decred/dcrd/chaincfg/v2 v2.2.0
	github.com/decred/dcrd/dcrutil/v2 v2.0.0
	github.com/decred/dcrd/wire v1.2.0
	github.com/decred/dcrwallet v1.2.2
	github.com/decred/dcrwallet/chain/v3 v3.0.0
	github.com/decred/dcrwallet/deployments/v2 v2.0.0 // indirect
	github.com/decred/dcrwallet/wallet/v3 v3.0.0
)

replace (
	github.com/decred/dcrd/fees/v2 => github.com/decred/dcrd/fees/v2 v2.0.0-20190817061140-8497b9843bcb

	github.com/decred/dcrwallet => github.com/jrick/btcwallet v0.0.0-20190903173710-02ab93ce28c3
	github.com/decred/dcrwallet/chain/v3 => github.com/jrick/btcwallet/chain/v3 v3.0.0-20190903173710-02ab93ce28c3
	github.com/decred/dcrwallet/deployments/v2 => github.com/jrick/btcwallet/deployments/v2 v2.0.0-20190903173710-02ab93ce28c3
	github.com/decred/dcrwallet/p2p/v2 => github.com/jrick/btcwallet/p2p/v2 v2.0.0-20190903173710-02ab93ce28c3
	github.com/decred/dcrwallet/rpc/client/dcrd => github.com/jrick/btcwallet/rpc/client/dcrd v0.0.0-20190903173710-02ab93ce28c3
	github.com/decred/dcrwallet/spv/v3 => github.com/jrick/btcwallet/spv/v3 v3.0.0-20190903173710-02ab93ce28c3
	github.com/decred/dcrwallet/ticketbuyer/v4 => github.com/jrick/btcwallet/ticketbuyer/v4 v4.0.0-20190903173710-02ab93ce28c3
	github.com/decred/dcrwallet/wallet/v3 => github.com/jrick/btcwallet/wallet/v3 v3.0.0-20190903173710-02ab93ce28c3
)
