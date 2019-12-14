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

- Run below command:

        go get github.com/0chain/gosdk
- Build the sample application sdkversion

        go build -o sdkversion sdkversion.go
- Run the executable

        ./sdkver
- If it prints the gosdk version installed then setup is complete.

### FAQ ###

- [How to install GO on any platform](https://golang.org/doc/install)
- [How to install different version of GO](https://golang.org/doc/install#extra_versions)
- [How to use go mod](https://blog.golang.org/using-go-modules)
