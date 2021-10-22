# 0chain/gosdk
The 0chain client SDK is written in GO.
This repository currently supports the following platforms:
- OSX Mojave 10.14.5
- LINUX (Ubuntu/bionic).
  - This includes all Ubuntu 18+ platforms, so Ubuntu 19, Linux Mint 19 etc. (apt based package installer)
- LINUX (RHEL/CENTOS 7+)
  - All Releases based on RHEL 7+, Centos 7+, Fedora 30 etc. (yum based package installer)

It is possible to support the sdk for other variations of Linux as well.

## Usage
- Save below code as sdkversion.go

        package main

        import (
            "fmt"

            "github.com/0chain/gosdk/zcncore"
        )

        func main() {
            fmt.Println("gosdk version: ", zcncore.GetVersion())
        }

- Run below command: (if you don't have gosdk already in your GOPATH)

        go get github.com/0chain/gosdk
- Build the sample application sdkversion

        go build -o sdkversion sdkversion.go
- Run the executable

        ./sdkversion
- If it prints the gosdk version installed then setup is complete.


## Mobile Builds (iOS and Android) ##
- gosdk can be build to use on Mobile platforms iOS and Android using gomobile.
- Xcode Command Line Tools is required to build SDK for iOS.
- Android studio with NDK is required to build SDK for Android
- Run below command for the first time to setup gomobile environment

        make setup-gomobile

- Use below commands in the root folder of the repo to build Mobile SDK

        For iOS and Android:
                make build-mobilesdk IOS=1 ANDROID=1
        For iOS only:
                make build-mobilesdk IOS=1
        For Android only:
                make build-mobilesdk ANDROID=1

## How to export a gosdk function to WebAssembly ##

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

3. Now you need to compile a new `<any_name>.wasm` (e.g. proxy.wasm). The command is currently: 

    ```bash
    $ GOOS=js CGO_ENABLED=0 GOARCH=wasm go -tags fullver build -o <any_name>.wasm github.com/0chain/gosdk/wasmsdk
    ```

4. You can compile a minimum version by adding tags `-minver`, for example:

    ```bash
    $ CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -tags minver -o min.wasm github.com/0chain/gosdk/wasmsdk
    ```

### An important note regarding export of a async function

If your golang function requires to be run asynchronously, you need to add more wrapper code where you are returning a Promise object.

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

## How to run unit test ##

### BLS unit test

It's advisable to put GOPATH as $TOP/../go, to avoid conflicts with this command: `go build ./...`

To run all the unit tests in `gosdk`: go test github.com/0chain/gosdk/zboxcore/sdk -v`
```bash
$ go test ./...
```

To run all the unit tests in `bls0chain_test.go`, run this command from $TOP: `go test github.com/0chain/gosdk/core/zcncrypto -v`

To run a specific unit test in `bls0chain_test.go` such as `TestSignatureScheme`, run this: `go test github.com/0chain/gosdk/core/zcncrypto -v -run TestSignatureScheme`

To run the coverage test in `gosdk:
```bash
$ go test <path_to_folder> -coverprofile=coverage.out
$ go tool cover -html=coverage.out
```

### WebAssembly

#### Using go test

1. You need to install nodejs first, see [this page](https://nodejs.org/en/download/) for further instructions

2. Add `/path/to/go/misc/wasm` to your `$PATH` environment variable (so that "go test" can find "go_js_wasm_exec"). For example in ubuntu, run `$export PATH=$PATH:/usr/local/go/misc/wasm/`.

3. You can then run the test by following the [BLS unit test](#bls-unit-test) above with adding the prefix environment `GOOS=js CGO_ENABLED=0 GOARCH=wasm` before `go test -v`.

#### Test in the client 

1. After you successfully [export the wasm package to proxy.wasm](#how-to-export-a-gosdk-function-to-webassembly), now you can test that `proxy.wasm`. 

2. We currently have a test page going at the js-client-sdk repo: `https://github.com/0chain/js-client-sdk/blob/gosdk/test/index.html`

3. You can replace the proxy.wasm at `https://github.com/0chain/js-client-sdk/blob/gosdk/test/proxy.wasm`

4. You need to startup a special test server in order to stream WASM files. Use this command from js-client-sdk $TOP: `sudo php -S localhost:82 test/server.php`

5. See "testethwallet" function in index.html for how the testing for ethwallet.go is done

6. To test the function you exported, it's probably as simple as calling "HelloWorld()". It should be a 1-liner.

### How to install `ffmepg` 

### On linux ubuntu

```
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
