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

## How to run BLS unit tests ##

It's advisable to put GOPATH as $TOP/../go, to avoid conflicts with this command: `go build ./...`

To run all the unit tests in `bls0chain_test.go`, run this command from $TOP: `go test github.com/0chain/gosdk/core/zcncrypto -v`

To run a specific unit test in `bls0chain_test.go` such as `TestSignatureScheme`, run this: `go test github.com/0chain/gosdk/core/zcncrypto -v -run TestSignatureScheme`

## How to export a gosdk function to proxy.wasm ##

Examples:
* `_sdkver/ethwallet.go` which exports the functions in `zcncore/ethwallet.go`.
* `_sdkver/wallet.go` which exports one function in `zcncore/wallet.go`.

Steps:

1a. If you are exporting a new function from `zcncore/wallet.go`, you should add to `_sdkver/wallet.go`

1b. If you are exporrting a function from a new file, you should create a new .go file for it, in the same style as `_sdkver/wallet.go` or `_sdkver/ethwallet.go`

2. In func main(), `https://github.com/0chain/gosdk/blob/jssdk/_sdkver/proxy.go`, you need to add this line:

```
	js.Global().Set("YOURFUNC", js.FuncOf(YOURFUNC))
```

3. Now you need to compile a new `proxy.wasm`. The command is currently: `GOOS=js GOARCH=wasm go build -o proxy.wasm proxy.go ethwallet.go wallet.go;`

3a. Please note that if you added a new golang file, then you need to add a new golang file to that compile command.

4. Now you need to test that `proxy.wasm`. We currently have a test page going at the js-client-sdk repo: `https://github.com/0chain/js-client-sdk/blob/gosdk/test/index.html`

4a. You can replace the proxy.wasm at `https://github.com/0chain/js-client-sdk/blob/gosdk/test/proxy.wasm`

4b. You need to startup a special test server in order to stream WASM files. Use this command from js-client-sdk $TOP: `sudo php -S localhost:82 test/server.php`

4c. See "testethwallet" function in index.html for how the testing for ethwallet.go is done

4d. To test the function you exported, it's probably as simple as calling "HelloWorld()". It should be a 1-liner.

### An important note regarding export of a async function

If your golang function requires to be run asynchronously, you need to add more wrapper code where you are returning a Promise object.

See "InitZCNSDK" example:

```
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


## FAQ ##

- [How to install GO on any platform](https://golang.org/doc/install)
- [How to install different version of GO](https://golang.org/doc/install#extra_versions)
- [How to use go mod](https://blog.golang.org/using-go-modules)
- [What is gomobile](https://godoc.org/golang.org/x/mobile/cmd/gomobile)
- [About XCode](https://developer.apple.com/xcode/)
- [Android Studio](https://developer.android.com/studio)
- [Android NDK](https://developer.android.com/ndk/)
