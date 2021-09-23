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


### Mobile Builds (iOS and Android) ###
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

### How to install `ffmepg` 

### On linux ubuntu

```
sudo apt-get install ffmpeg
sudo apt-get install v4l-utils
``

### FAQ ###

- [How to install GO on any platform](https://golang.org/doc/install)
- [How to install different version of GO](https://golang.org/doc/install#extra_versions)
- [How to use go mod](https://blog.golang.org/using-go-modules)
- [What is gomobile](https://godoc.org/golang.org/x/mobile/cmd/gomobile)
- [About XCode](https://developer.apple.com/xcode/)
- [Android Studio](https://developer.android.com/studio)
- [Android NDK](https://developer.android.com/ndk/)
