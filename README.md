# TradeBlocks [![Build Status](https://travis-ci.com/jephir/tradeblocks.svg?token=H5s5urysT233MRnGw5EA&branch=master)](https://travis-ci.com/jephir/tradeblocks)

Decentralized exchange implementation for Binance Dexathon.

## Installation

```sh
$ go install -i github.com/jephir/tradeblocks/cmd/tradeblocks github.com/jephir/tradeblocks/cmd/tradeblocks-node
```

## Demo

Create a user `alice` and a new token `my-coin` with 1000 tokens.

```sh
$ XTB_ALICE="$(tradeblocks register alice)"
$ XTB_MY_COIN="$(tradeblocks register my-coin)"
$ tradeblocks issue 1000
```

Send 50 tokens to `alice`.

```sh
$ SEND_TX="$(tradeblocks send $XTB_ALICE $XTB_MY_COIN 50)"
$ tradeblocks login alice
$ tradeblocks receive $SEND_TX
```

## Running Tests

```sh
$ go test -v ./...
```

## Authors

* Julian Hoang
* Eric Parker
