# 0chain/gosdk

The Züs client SDK written in Go programming language.

- [GoSDK - a Go based SDK for Züs dStorage]()
  - [Züs Overview ](#overview)
  - [Installation](#installation)
  - [Mobile Builds(iOS and Android)](#mobile-builds)
  - [Exposing a gosdk function to mobilesdk](#expose-a-gosdk-function-to-mobilesdk)
  - [Export a gosdk function to WebAssembly](#export-a-gosdk-function-to-webassembly)
  - [Running Unit Tests](#running-unit-tests)
  - [FAQ](#faq)

## Overview

[Züs](https://zus.network/) is a high-performance cloud on a fast blockchain offering privacy and configurable uptime. It is an alternative to traditional cloud S3 and has shown better performance on a test network due to its parallel data architecture. The technology uses erasure code to distribute the data between data and parity servers. Züs storage is configurable to provide flexibility for IT managers to design for desired security and uptime, and can design a hybrid or a multi-cloud architecture with a few clicks using [Blimp's](https://blimp.software/) workflow, and can change redundancy and providers on the fly.

For instance, the user can start with 10 data and 5 parity providers and select where they are located globally, and later decide to add a provider on-the-fly to increase resilience, performance, or switch to a lower cost provider.

Users can also add their own servers to the network to operate in a hybrid cloud architecture. Such flexibility allows the user to improve their regulatory, content distribution, and security requirements with a true multi-cloud architecture. Users can also construct a private cloud with all of their own servers rented across the globe to have a better content distribution, highly available network, higher performance, and lower cost.

[The QoS protocol](https://medium.com/0chain/qos-protocol-weekly-debrief-april-12-2023-44524924381f) is time-based where the blockchain challenges a provider on a file that the provider must respond within a certain time based on its size to pass. This forces the provider to have a good server and data center performance to earn rewards and income.

The [privacy protocol](https://zus.network/build) from Züs is unique where a user can easily share their encrypted data with their business partners, friends, and family through a proxy key sharing protocol, where the key is given to the providers, and they re-encrypt the data using the proxy key so that only the recipient can decrypt it with their private key.

Züs has ecosystem apps to encourage traditional storage consumption such as [Blimp](https://blimp.software/), a S3 server and cloud migration platform, and [Vult](https://vult.network/), a personal cloud app to store encrypted data and share privately with friends and family, and [Chalk](https://chalk.software/), a high-performance story-telling storage solution for NFT artists.

Other apps are [Bolt](https://bolt.holdings/), a wallet that is very secure with air-gapped 2FA split-key protocol to prevent hacks from compromising your digital assets, and it enables you to stake and earn from the storage providers; [Atlus](https://atlus.cloud/), a blockchain explorer and [Chimney](https://demo.chimney.software/), which allows anyone to join the network and earn using their server or by just renting one, with no prior knowledge required.

## Installation

### Supported Platforms

This repository currently supports the following platforms:

- Mac OSX Mojave 10.14.5 or Above
- Linux (Ubuntu/bionic): This includes all Ubuntu 18+ platforms, so Ubuntu 19, Linux Mint 19 etc. (apt based package installer)
- Linux (RHEL/CENTOS 7+): All Releases based on RHEL 7+, Centos 7+, Fedora 30 etc. (yum based package installer)

### Instructions

- Go is required to build gosdk code. Instructions can be found [here](https://go.dev/doc/install)

1.  Save below code as `sdkversion.go`

        package main

        import (
            "fmt"

            "github.com/0chain/gosdk/zcncore"
        )

        func main() {
            fmt.Println("gosdk version: ", zcncore.GetVersion())
        }

2.  Run below command: (if you don't have gosdk already in your GOPATH)

        go get github.com/0chain/gosdk

3.  Build the sample application sdkversion

        go build -o sdkversion sdkversion.go

4.  Run the executable

        ./sdkversion

5.  If it prints the gosdk version installed then setup is complete.

## Mobile Builds

- gosdk can be built for iOS and Android using gomobile.
- Xcode Command Line Tools is required to build the SDK for iOS.
- Android studio with NDK is required to build the SDK for Android.
- See [FAQ](#faq) for installing Go, gomobile Xcode or Android Studio.

Steps:

1.  Run the command below for the first time to setup the gomobile environment:

        make setup-gomobile

2.  In case the Go package is not found in `golang.org/x/mobile/bind`, run:
    `go get golang.org/x/mobile/bind`
3.  Run below commands in the root folder of the repo to build the Mobile SDK:

        For iOS only:
                make build-ios
        For Android only:
                make build-android

## Expose a gosdk function to mobilesdk

Examples:

- `mobilesdk/sdk/common.go`, which exports the functions in `core/encryption/hash.go`.

Steps:

1. If you are exposing:

   - a new function from an existing file, such as `zboxcore/sdk/allocation.go`, you should add a function to `mobilesdksdk/zbox/allocation.go`. This new function should call the gosdk function you intend to expose.
   - a function from a new file, you should create a new `<filename>.go` file for it. This should follow the same style as `mobilesdksdk/zbox/allocation.go`. In the new file, call the gosdk function you intend to expose.

2. Build the Mobile SDK as mentioned in the 'Mobile Builds' section of this file to build the aar file used in the mobile application you are developing.

## Export a gosdk function to WebAssembly

Examples:

- `wasmsdk/ethwallet.go` which exports the functions in `zcncore/ethwallet.go`.
- `wasmsdk/wallet.go` which exports one function in `zcncore/wallet.go`.

Steps:

1. If you are exporting:

   - a new function from `zcncore/wallet.go`, you should add to `wasmsdk/wallet.go`

   - a function from a new file, you should create a new `<filename>.go` file for it, in the same style as `wasmsdk/wallet.go` or `wasmsdk/ethwallet.go`

2. In func main(), `https://github.com/0chain/gosdk/wasmsdk/proxy.go`, you need to add this line:

   ```golang
       js.Global().Set("YOURFUNC", js.FuncOf(YOURFUNC))
   ```

3. Now you need to compile a new `<any_name>.wasm` (e.g. proxy.wasm). Currently, the right version to compile wasm is with Go version 1.16. So make sure you have it to make the wasm build works properly. In order to compile, run the following command:

   ```bash
   $ GOOS=js CGO_ENABLED=0 GOARCH=wasm go build -o <any_name>.wasm github.com/0chain/gosdk/wasmsdk
   ```

### An important note regarding export of an async function

If your golang function needs to suport asynchronous execution, you need to add more wrapper code where you are returning a Promise object.

See "InitZCNSDK" example:

```golang
func InitZCNSDK(this js.Value, p []js.Value) interface{} {
	blockWorker := p[0].String()
	signscheme := p[1].String()

	handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		// reject := args[1]

		go func() {
			err := zcncore.InitZCNSDK(blockWorker, signscheme)
			if err != nil {
				fmt.Println("error:", err)
			}
			resolve.Invoke()
		}()

		return nil
	})

	// Create and return the Promise object
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}
```

## Running Unit Tests

### BLS unit test

It's advisable to put GOPATH as `$TOP/../go`, to avoid conflicts with this command: `go build ./...`

To run all the unit tests in `gosdk`: `go test github.com/0chain/gosdk/zboxcore/sdk -v`

```bash
$ go test ./...
```

To run all the unit tests in `bls0chain_test.go`, run this command from $TOP: `go test github.com/0chain/gosdk/core/zcncrypto -v`

To run a specific unit test in `bls0chain_test.go`, such as `TestSignatureScheme`, run: `go test github.com/0chain/gosdk/core/zcncrypto -v -run TestSignatureScheme`

To run the coverage test in `gosdk`:

```bash
$ go test <path_to_folder> -coverprofile=coverage.out
$ go tool cover -html=coverage.out
```

### WebAssembly

#### Using go test

1. You need to install nodejs first, see [this page](https://nodejs.org/en/download/) for further instructions

2. Add `/path/to/go/misc/wasm` to your `$PATH` environment variable (so that `go test` can find `go_js_wasm_exec`). For example in Ubuntu, run `$export PATH=$PATH:/usr/local/go/misc/wasm/`.

3. You can then run the test by following the [BLS unit test](#bls-unit-test) above by adding the prefix environment `GOOS=js CGO_ENABLED=0 GOARCH=wasm`:
   ```bash
   go test -tags test -v github.com/0chain/gosdk/wasmsdk
   ```

#### Test in the client

1. After you successfully [export the wasm package to proxy.wasm](#how-to-export-a-gosdk-function-to-webassembly), you can test the exported `proxy.wasm`.

2. We currently have a test page going at the js-client-sdk repo: `https://github.com/0chain/js-client-sdk/blob/gosdk/test/index.html`

3. You can replace the proxy.wasm at `https://github.com/0chain/js-client-sdk/blob/gosdk/test/proxy.wasm`

4. You need to start a special test server in order to stream WASM files. Use the following command from js-client-sdk $TOP: `sudo php -S localhost:82 test/server.php`

5. See "testethwallet" function in index.html for how the testing for ethwallet.go is done.

6. To test the function you exported, it's probably as simple as calling "HelloWorld()". It should be a 1-liner.

### How to install `ffmpeg`

#### On Ubuntu Linux

```bash
sudo apt-get install ffmpeg
sudo apt-get install v4l-utils
```

### FAQ

- [How to install GO on any platform](https://golang.org/doc/install)
- [How to install different version of GO](https://golang.org/doc/install#extra_versions)
- [How to use go mod](https://blog.golang.org/using-go-modules)
- [What is gomobile](https://godoc.org/golang.org/x/mobile/cmd/gomobile)
- [About XCode](https://developer.apple.com/xcode/)
- [Android Studio](https://developer.android.com/studio)
- [Android NDK](https://developer.android.com/ndk/)
