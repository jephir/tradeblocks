# TradeBlocks [![Build Status](https://travis-ci.com/jephir/tradeblocks.svg?token=H5s5urysT233MRnGw5EA&branch=master)](https://travis-ci.com/jephir/tradeblocks)

Decentralized exchange implementation for Binance Dexathon.

## Summary

**TradeBlocks** is a decentralized token exchange network that provides near-instant token trading with high scalability. This is achieved by utilizing a separate blockchain for each account-token pair in the network. To transfer tokens from one account to another, the sender creates a send transaction on their own blockchain, and the receiver creates a receive transaction on their own blockchain. This makes token transfers asynchronous and massively increases the throughput of the network. To trade one type of token for a different type, an initiator first sends tokens into a swap blockchain and creates an offer transaction. Next, a counterparty sends tokens into the swap blockchain and creates a commit transaction. Finally, the parties create receive transactions on their own account-token blockchains to receive the swapped tokens. If the counterparty doesn’t send tokens for the swap, the initiator can create a refund transaction to return their tokens back to their own account-token blockchain. If a fork occurs, the network uses a delegated proof-of-stake protocol to resolve the conflict.

## Installation

```sh
$ go install -i github.com/jephir/tradeblocks/cmd/tradeblocks
```

## Demo

1.  Start the node server.

```sh
$ tradeblocks node
```

2.  Create a user `alice` and a new token `apple-coin` with 1000 tokens.

```sh
$ XTB_ALICE="$(tradeblocks register alice)"
$ XTB_APPLE_COIN="$(tradeblocks register apple-coin)"
$ tradeblocks login apple-coin
$ tradeblocks issue 1000
```

3.  Send 50 `apple-coin` tokens to `alice`.

```sh
$ XTB_SEND="$(tradeblocks send $XTB_ALICE $XTB_APPLE_COIN 50)"
$ tradeblocks login alice
$ tradeblocks open $XTB_SEND
```

4.  Create a new token `banana-coin` with 2000 tokens. Then, offer to trade 25 `banana-coin` for 50 `apple-coin` with `alice`. Finally, accept the trade as `alice`.

```sh
$ XTB_BANANA_COIN="$(tradeblocks register banana-coin)"
$ XTB_TRADE="$(tradeblocks trade $XTB_BANANA_COIN 25 $XTB_ALICE $XTB_APPLE_COIN 50)"
$ tradeblocks login alice
$ tradeblocks trade $XTB_TRADE
```

## Running Tests

```sh
$ go test -v ./...
```

## Web Servers

Need both of the following running:

1.  Blockchain server

```
$ tradeblocks node
```

2.  React server

```
$ cd web
$ npm start
```

## Authors

- Julian Hoang
- Eric Parker
