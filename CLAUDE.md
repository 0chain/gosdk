# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build Go SDK (requires bn256 tag)
make gosdk-build

# Run all unit tests (sequential execution required)
make gosdk-test
# Or directly:
go test -tags bn256 -p 1 ./...

# Run a single test
go test -tags bn256 -v -run TestName ./path/to/package

# Generate mocks (uses mockery, outputs to mocks/ subdirectories)
make gosdk-mocks

# Lint code
make lint          # Skips wasmsdk
make lint-wasm     # WASM only

# Build WASM binary
make wasm-build

# Build mobile SDKs
make build-ios
make build-android

# Vendor dependencies
make vendor
```

## Dependencies

The `bn256` build tag requires the **Herumi BLS/MCL** native libraries. On Ubuntu:
```bash
make install-herumi-ubuntu
```
This installs `libmclbn256.so` and related BLS libraries to `/usr/local/lib`.

## Architecture Overview

This is the **Züs (0chain) Go SDK** - a distributed storage SDK for the Züs blockchain network. The SDK supports multiple platforms: native Go, WebAssembly, iOS, Android, and Windows.

### Core Package Structure

**`zcncore/`** - Blockchain wallet operations
- Wallet management (Ethereum, multi-sig, 0chain wallets)
- Transaction issuance and queries
- Network interaction with miners/sharders

**`zboxcore/`** - Storage operations (primary SDK)
- `sdk/` - Core storage operations (allocation, upload, download, repair)
- `allocationchange/` - Change tracking for allocations
- `fileref/` - File reference management
- `marker/` - WriteMarker proofs for commitment
- `encoder/` - Erasure coding (Reed-Solomon)

**`core/`** - Foundation utilities
- `zcncrypto/` - BLS signatures, ED25519 (uses Herumi library)
- `transaction/` - Transaction handling and signing
- `resty/` - HTTP client wrapper (abstracted for WASM compatibility)
- `node/` - Node management and caching
- `sys/` - File system abstractions (disk and in-memory)
- `conf/` - Configuration loading from `~/.zcn/config.yaml` via Viper

**`zcnbridge/`** - Token bridge between 0chain and Ethereum

**Platform SDKs:**
- `wasmsdk/` - WebAssembly exports via `jsbridge/`
- `mobilesdk/` - iOS/Android wrappers via gomobile
- `winsdk/` - Windows DLL (C-shared)

### SDK Configuration

Config is loaded from `~/.zcn/config.yaml`. Key fields:

```yaml
block_worker: http://<0dns-host>:9091  # Required: 0DNS network API URL
signature_scheme: bls0chain
min_submit: 50          # % of blobbers to submit to (default 20, max 100)
min_confirmation: 50    # % of sharders for tx confirmation (default 10, max 100)
confirmation_chain_length: 3
max_txn_query: 5
query_sleep_time: 5
sharder_consensous: 3   # Min sharders for SCRestAPI consensus (default 3)
verify_optimistic: true # Use optimistic transaction verification
```

The `core/conf` package exposes `LoadConfigFile()` and `LoadConfig()`. All percentage values are clamped to [1, 100].

### Key Architectural Patterns

1. **Uint128 bitmask consensus**: Blobber participation is tracked with a `Uint128` bitmask (supports up to 128 blobbers). `CountOnes()` counts successful blobbers; `operationMask` in `MultiOperation` tracks which blobbers participated per operation. See `zboxcore/zboxutil/uint128.go`.

2. **Weighted node health scoring**: `NodeHolder` in `core/client/node.go` tracks a 20-entry sliding window of success/failure per node. `HealthyByLFB()` queries all sharders for the current round (LFB = Last Finalized Block) and only uses nodes within **3 blocks** of the highest LFB.

3. **Chunked uploads with erasure coding**: Files are chunked, Reed-Solomon encoded across `DataShards + ParityShards` blobbers, then committed via WriteMarker consensus.

4. **WriteMarker system**: Commits use V2 markers with weighted merkle trie proofs (`wmpt.WeightedMerkleTrie`). Each blobber must sign the marker before the allocation is committed.

5. **Channel-based commit queuing**: `commitChan` map in `zboxcore/sdk/` queues commit operations per allocation to prevent concurrent conflicts.

6. **SCRestAPI consensus voting**: `core/client/http.go` uses status code dominance voting across sharders (25% threshold; ties break to 200 OK).

7. **Platform abstraction**: Build tags (`bn256`, `js`, `wasm`) enable platform-specific implementations.

### Build Tags

- `bn256` - Required for cryptographic operations (always use for builds/tests)
- Platform-specific tags handle WASM, mobile, and Windows builds automatically

### Test Patterns

Tests use `testify/require` and `testify/assert`. Run tests sequentially (`-p 1`) due to shared state in some tests. Mocks are generated via `mockery` and live in `mocks/` subdirectories within each package.

### Key Files

- `zboxcore/sdk/allocation.go` - Main allocation management (largest file, ~115KB)
- `zboxcore/sdk/consensus.go` - Thread-safe consensus counter (`Consensus` struct)
- `core/conf/config.go` - Configuration structure and loading
- `core/client/node.go` - Node health tracking and LFB-based selection
- `core/transaction/entity.go` - Transaction struct and smart contract addresses
- `core/zcncrypto/` - Cryptographic operations requiring Herumi BLS library
- `constants/` - Shared constants and error definitions

## Related Docs

- [NFS_INTEGRATION.md](./NFS_INTEGRATION.md) — How the zs3server NFS gateway uses GoSDK allocation methods
