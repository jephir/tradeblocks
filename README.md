# TradeBlocks [![Build Status](https://travis-ci.com/jephir/tradeblocks.svg?token=H5s5urysT233MRnGw5EA&branch=master)](https://travis-ci.com/jephir/tradeblocks)

Decentralized exchange proof of concept for Binance Dexathon.

## Presentation Video

[![Presentation Video](https://img.youtube.com/vi/aNVp_qxYUbs/0.jpg)](https://www.youtube.com/watch?v=aNVp_qxYUbs)

## [Whitepaper](TradeBlocksWhitepaper.pdf)

## Summary

**TradeBlocks** is a decentralized token exchange network that provides near-instant token trading with high scalability.

This is achieved by utilizing a separate blockchain for each account-token pair in the network. To transfer tokens from one account to another, the sender creates a send transaction on their own blockchain, and the receiver creates a receive transaction on their own blockchain. This makes token transfers asynchronous and massively increases the throughput of the network. 

To trade one type of token for a different type, an initiator first sends tokens into a swap blockchain and creates an offer transaction. Next, a counterparty sends tokens into the swap blockchain and creates a commit transaction. Finally, the parties create receive transactions on their own account-token blockchains to receive the swapped tokens. If the counterparty doesn’t send tokens for the swap, the initiator can create a refund transaction to return their tokens back to their own account-token blockchain. 

If a fork occurs, the network uses a delegated proof-of-stake protocol to resolve the conflict.

## Requirements

- [Go 1.10](https://golang.org/)
- [GCC 7.3](https://gcc.gnu.org/)
- [Node 6.11.5](https://nodejs.org/en/)

## Installation

```sh
$ go install -i github.com/jephir/tradeblocks/cmd/tradeblocks
```

## Demo

To view created blocks, launch the web server (see section below).

### Limit Orders

1.  Start the node server.

```sh
$ tradeblocks node
```

2.  Create tokens `t1` and `t2` with 1000 tokens each.

```sh
$ XTB_T1="$(tradeblocks register t1)"
$ XTB_T2="$(tradeblocks register t2)"
$ tradeblocks login t1
$ tradeblocks issue 1000
$ tradeblocks login t2
$ tradeblocks issue 1000
```

3.  Create a sell order for 100 units of `t2` token for `t1` token at 2 price per unit (200 `t1`).

```sh
$ tradeblocks sell 100 $XTB_T2 2 $XTB_T1
```

4.  Create a matching buy order. The node will then execute the swap.

```sh
$ tradeblocks login t1
$ tradeblocks buy 100 $XTB_T2 2 $XTB_T1
```

### Low-Level Transactions

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
$ XTB_SEND1="$(tradeblocks send $XTB_ALICE $XTB_APPLE_COIN 50)"
$ tradeblocks login alice
$ tradeblocks open $XTB_SEND1
```

4.  Create a new token `banana-coin` with 2000 tokens.

```sh
$ XTB_BANANA_COIN="$(tradeblocks register banana-coin)"
$ tradeblocks login banana-coin
$ tradeblocks issue 2000
```

5.  Create an offer to trade 25 `banana-coin` for 50 `apple-coin` with `alice`.

```sh
$ XTB_OFFER_ID = "BANANA_APPLE_OFFER"
$ XTB_OFFER_LINK = $XTB_OFFER_ID += ":offer:"
$ XTB_OFFER_LINK += $XTB_BANANA_COIN
$ XTB_SEND2 = "$(tradeblocks send $XTB_OFFER_LINK $XTB_BANANA_COIN 25)"
$ XTB_CREATE_ORDER ="$(tradeblocks create-order $XTB_SEND2 $XTB_OFFER_ID false $XTB_APPLE_COIN 2)"
```

6.  Create the offer swap for `alice` to accept the order

```sh
$ tradeblocks login alice
$ XTB_SWAP_ID = "BANANA_APPLE_SWAP"
$ XTB_SWAP_LINK = $XTB_SWAP_ID += ":swap:"
$ XTB_SWAP_LINK += $XTB_BANANA_COIN
$ XTB_SEND3 = "$(tradeblocks send $XTB_SWAP_LINK $XTB_APPLE_COIN 50)"
$ XTB_SWAP_OFFER = "$(tradeblocks offer $XTB_SEND3 $XTB_SWAP_ID $XTB_BANANA_COIN $XTB_BANANA_COIN 25)"
```

7.  Accept the swap for `banana-coin`

```sh
$ XTB_ACCEPT_ORDER = "$(tradeblocks accept $XTB_CREATE_ORDER $XTB_SWAP_OFFER)"
$ XTB_SWAP_COMMIT = "$(tradeblocks commit $XTB_SWAP_OFFER $SEND2)"
```

8.  Receive the coins

```sh
$ XTB_RECEIVE1 = "$(tradeblocks receive $XTB_SWAP_COMMIT)"
$ tradeblocks login banana-coin
$ XTB_RECEIVE1 = "$(tradeblocks receive $XTB_SWAP_COMMIT)"
```

## Commands

* `tradeblocks node -listen <address> -bootstrap <url> -dir <path>`
  * Start a new node server on this machine
* `tradeblocks register <name>`
  * Register a new key pair
* `tradeblocks login <name>`
  * Login to an existing key pair
* `tradeblocks issue <balance>`
  * Issue new tokens
* `tradeblocks send <address> <token> <amount>`
  * Send tokens to an address
* `tradeblocks open <block>`
  * Open a new account from a send
* `tradeblocks open-from-swap <block>`
  * Open a new account from a swap
* `tradeblocks receive <block>`
  * Receive tokens from a send
* `tradeblocks offer <send> <id> <counterparty> <base> <quantity> <executor> <fee>`
  * Offer a swap with a counterparty
* `tradeblocks commit <offer> <send>`
  * Commit a swap as a counterparty
* `tradeblocks refund-left <offer>`
  * Cancel a swap as the initiator
* `tradeblocks refund-right <refund-left>`
  * Refund yourself as a counterparty
* `tradeblocks create-order <send> <id> <partial> <quote> <price> <executor> <fee>`
  * Create a new order
* `tradeblocks accept-order <swap> <link>`
  * Accept an incoming swap for your order
* `tradeblocks refund-order <order>`
  * Cancel an order
* `tradeblocks sell <quantity> <base> <ppu> <quote>`
  * Create a limit sell order
* `tradeblocks buy <quantity> <base> <ppu> <quote>`
  * Create a limit buy order
* `tradeblocks cat <hash>`
  * Print out a block

## Running Tests

```sh
$ go get ./...
$ go test -v ./...
```

## Web Servers

Requires both of the following running:

1.  Blockchain server

```
$ tradeblocks node
```

2.  React server

```
$ cd web
$ npm install
$ npm start
```

## Known Issues

* Nodes do not sync all blocks from a connecting node to a root node
* A limit sell order must be placed before a matching limit buy order
* Conflicted blocks are not always removed from the block store

## Authors

- Julian Hoang <julian.b.hoang@gmail.com>
- Eric Parker <eric.parker13@gmail.com>
