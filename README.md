# TradeBlocks [![Build Status](https://travis-ci.com/jephir/tradeblocks.svg?token=H5s5urysT233MRnGw5EA&branch=master)](https://travis-ci.com/jephir/tradeblocks)

Decentralized exchange implementation for Binance Dexathon.

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

* Julian Hoang
* Eric Parker
