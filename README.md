# 0chain/gosdk

The Züs client SDK written in Go programming language.

- [GoSDK - a Go based SDK for Züs dStorage]()
  - [Züs Overview ](#overview)
  - [Installation](#installation)
  - [Mobile Builds(iOS and Android)](#mobile-builds)
  - [SDK Reference](#sdk-reference)
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
 - Linux (Ubuntu Preferred) Version: 20.04 and Above
 - Mac(Apple Silicon or Intel) Version: Big Sur and Above
 - Linux (RHEL/CENTOS 7+): All Releases based on RHEL 7+, Centos 7+, Fedora 30 etc. (yum based package installer)

 ### Required Software Dependencies
  - Go is required to build gosdk code. Instructions can be found [here](https://github.com/0chain/0chain/blob/hm90121-patch-1/standalone_guides.md#install-go).

 ### Instructions
 1. Open terminal and clone the gosdk repo.
  ```
  git clone https://github.com/0chain/gosdk.git
  ```

 2. Move to gosdk directory using `cd gosdk` command and save the code below as a file named `sdkversion.go`.

        package main

        import (
            "fmt"

            "github.com/0chain/gosdk/zcncore"
        )

        func main() {
            fmt.Println("gosdk version: ", zcncore.GetVersion())
        }

3. Run the command below to retrieve the gosdk package (if you don't have gosdk already in your GOPATH)

        go get github.com/0chain/gosdk
4. Build the sample application `sdkversion.go` with gosdk using the command below: 

        go build -o sdkversion sdkversion.go
5. Run the executable using the command below:

        ./sdkversion
6. If the executable prints the installed version of the Go SDK, then the setup is complete.

      Sample Output:
      ```
     gosdk version:  v1.8.17-78-g80b63345
      ```
 
## Mobile Builds

GoSDK can be built to use on mobile platforms iOS and Android using gomobile. To setup gomobile environment :
 
1. Make sure $HOME/go/bin/ is in your $PATH. 

```
export PATH=${PATH}:$HOME/go/bin/
```
Note:Edit your bash profile. Your profile is a file in your home directory named either `.profile` or `.bash_profile`. Add the line above to the file for setting `$GOPATH` system wide.

2. Now to install and intialize gomobile package run the commands below:

```
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

3. Run `gomobile` command to check whether its installed.

    Sample Output:

```
Gomobile is a tool for building and running mobile apps written in Go.
For detailed instructions, see https://golang.org/wiki/Mobile.

Usage:

        gomobile command [arguments]

Commands:

        bind        build a library for Android and iOS
        build       compile android APK and iOS app
        clean       remove object files and cached gomobile files
        init        build OpenAL for Android
        install     compile android APK and install on device
        version     print version

Use 'gomobile help [command]' for more information about that command.
```

4. After successfully installing Go Mobile, proceed to compile the Mobile SDK.

     For Linux : Download and Install [Android studio](https://developer.android.com/studio) with [NDK](https://developer.android.com/ndk/guides#download-ndk) and [JDK](https://www.oracle.com/java/technologies/downloads/) to build SDK for 
                 Android.

     For Mac: Download and Install [Xcode Command Line Tools](https://developer.apple.com/download/all/) to build SDK for iOS. Requires Apple ID to download.

     Commands for building mobilesdk are below :
     ```
     For iOS only:       make build-ios  
     For Android only:     make build-android  
     ```

     Sample Response for android :
     ```
     jar: zcncore/Zcncore.java
     C:/Users/0chain/Desktop/gosdk/out/androidsdk/zcncore.aar. - [OK]
     ```

     Sample Response for iOS:
    ```
    xcframework successfully written out to: 
    /Users//gosdk/out/iossdk/ios/zcncore.xcframework   
    /Users/0chain/gosdk/out/iossdk/ios/zcncore.xcframework. - [OK]
    ```

   The response will successfully build a library file which includes everything needed to build a Züs mobile app, including source code and resource files. Now you can use the library into your own iOS or Android project.

## SDK Reference

This section includes links to all packages in gosdk. Refer to the [gosdk package](https://pkg.go.dev/github.com/0chain/gosdk) for a comprehensive list of all packages in the Go SDK.

Here is an overview of all modules mentioned in the [gosdk package](https://pkg.go.dev/github.com/0chain/gosdk#section-directories) . 

**Constants:**
The Constants package serves as a repository for constants. The naming convention adheres to the use of MixedCaps or mixedCaps, favoring them over underscores when creating multiword names.

**Core:**
The Core module forms the essential foundation, providing fundamental functionalities and serving as the backbone for various components.

**Dev:**
The Dev package equips developers with tools tailored for local development, fostering a seamless environment for building and testing applications.

**Errors Module:**
The Errors Module encompasses a comprehensive set of tools and utilities dedicated to handling errors efficiently, ensuring robust error management within the system.

**Mobile SDK:**
The Mobile SDK facilitates the development of mobile applications, offering a set of software development tools specifically designed for mobile platforms.

**SDKs:**
The SDKs module encompasses various software development kits, each tailored to specific purposes, contributing to a versatile and comprehensive development ecosystem.

**Wasm SDK:**
The Wasm SDK focuses on providing tools and resources for the WebAssembly (Wasm) development environment, empowering developers to harness the potential of this emerging technology.

**Win SDK:**
The Win SDK caters to Windows application development, delivering a suite of tools and resources optimized for creating software on the Windows operating system.

**zbox API:**
The zbox API module serves as the interface for the zbox system, offering a set of functions and protocols for seamless communication and integration.

**zbox Core:**
The zbox core constitutes the foundational core of the zbox architecture, providing the essential framework for its various components.

**ZCN Bridge:**
The ZCN bridge module facilitates the connection and interaction between the Züs (ZCN) network and external systems, ensuring smooth interoperability.

**ZCN Core:**
ZCN Core represents the heart of the Züs (ZCN) network, encapsulating the essential components and functionalities that drive its operations.

**ZCN Swap:**
The ZCN Swap module focuses on enabling secure and efficient swapping of assets within the ZeroChain (ZCN) ecosystem.

**zmagma core:**
The zmagma Core encompasses the core functionalities and components of the Magma system, contributing to the overall stability and performance of the system.

**znft:**
The ZNFT module is dedicated to Non-Fungible Tokens (NFTs), providing tools and resources for the creation, management, and interaction with NFTs within the system.

## Expose a gosdk function to mobilesdk 
Examples:
* `mobilesdk/sdk/common.go`, which exports the functions in `core/encryption/hash.go`.

Steps:

1. If you are exposing:

    - a new function from an existing file, such as `zboxcore/sdk/allocation.go`, you should add a function to `mobilesdksdk/zbox/allocation.go`. This new function should call the gosdk function you intend to expose.
    - a function from a new file, you should create a new `<filename>.go` file for it. This should follow the same style as `mobilesdksdk/zbox/allocation.go`. In the new file, call the gosdk function you intend to expose.

2. Build the Mobile SDK as mentioned in the 'Mobile Builds' section of this file to build the aar file used in the mobile application you are developing.

## Export a gosdk function to WebAssembly 

Examples:
* `wasmsdk/ethwallet.go` which exports the functions in `zcncore/ethwallet.go`.
* `wasmsdk/wallet.go` which exports one function in `zcncore/wallet.go`.

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

### FAQ ###

- [How to install GO on any platform](https://golang.org/doc/install)
- [How to install different version of GO](https://golang.org/doc/install#extra_versions)
- [How to use go mod](https://blog.golang.org/using-go-modules)
- [What is gomobile](https://godoc.org/golang.org/x/mobile/cmd/gomobile)
- [About XCode](https://developer.apple.com/xcode/)
- [Android Studio](https://developer.android.com/studio)
- [Android NDK](https://developer.android.com/ndk/)
