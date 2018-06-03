# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.3.0 - 2018-06-03

### Added

* Swap chain validation and tests
* Order chain validation and tests
* Block conflict detection
* Vote (ordering) blocks

### Changed

* Block store account validation enabled
* Unit tests now check for signed blocks

## 0.2.0 - 2018-05-27

### Added

* `node` CLI command
* Filesystem storage for block store
* Block graph bootstrap from a root node
* Push-based multi-node block synchronization
* Signing and verification of account blocks
* SSE endpoint for retrieving account information
* Visual block explorer at http://localhost:8080

### Changed

* `receive` block validation disallows duplicate `send` claims

## 0.1.0 - 2018-05-17

### Added

* `register`, `login`, `issue`, `open`, `send`, and `receive` CLI commands
* Account, swap, and order blocks
* Account creation using public/private key pair
* Creating new tokens on account chain
* Sending and receiving tokens on account chain
* Graph validation
