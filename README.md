# 0chain/gosdk
The 0chain client SDK is written in GO. 
This repository currently supports the following platforms:
- OSX Mojave 10.14.5 
- LINUX (Ubuntu/bionic).
- - This includes all Ubuntu 18+ platforms, so Ubuntu 19, Linux Mint 19 etc. (apt based package installer)
- ADDED
- LINUX (RHEL/CENTOS 7+)
- - All Releases based on RHEL 7+, Centos 7+, Fedora 30 etc. (yum based package installer)

See Step-by-Step Installation guide plus Deployment Guide at bottom

It is possible to support the sdk for other variations of Linux as well. 

## Build and Installation 
0chain/gosdk is build and installed using [GNU Make](https://www.gnu.org/software/make/). 
The Makefile has following targets to ease installation of build tools and the 0chain/gosdk. 

**Success of installation of the library and GO modules is highly dependent upon the prior installed
packages on the developer system.**

These steps have been tested out thoroughly on OSX Mojave 10.14.5 and Vanilla Ubuntu BIONIC. 

0chain/gosdk has heavy dependency on [BLS](https://github.com/herumi/bls) and [MCL](https://github.com/herumi/mcl) 
provided by [MITSUNARI Shigeo](https://github.com/herumi). Developers should refer to those links when they encounter any errors. 

Please send email to [alphanet@0chain.net](mailto:alphanet@0chain.net) if you encounter any problems.

|TARGET       |Description   |
|:----        |:----------   |
| build-tools | Install go, jq and supporting tools|
| install     | Install herumi and gosdk|
| install-herumi |Build, Test and Install the herumi library|
| install-gosdk | Build and test 0chain gosdk modules|
| clean         | Delete all the build output |


### FAQ ###

- [How to install GO on any platform](https://golang.org/doc/install)
- [How to install different version of GO](https://golang.org/doc/install#extra_versions)
- How to install build tools on linux
```
        sudo apt-get install build-essential
``` 

- [What are the tools installed by build-tools on darwin](./_util/build_darwin.mk)
- [What are the tools installed by build-tools on linux](./_util/build_linux.mk)

- [What is snap ?](https://docs.snapcraft.io/getting-started)

- Will sudo apt-get install still work ?
  Ubuntu bionic has moved several packages to use snap. Some packages can still be downloaded
  the apt-get method. 

- How to install GO on ubuntu using snap
```.env
        sudo snap install go --classic
```
- [What are GO modules](https://github.com/golang/go/wiki/Modules)
- [How to: Install Go 1.11.2 on Ubuntu using snap](https://medium.com/@patdhlk/how-to-install-go-1-9-1-on-ubuntu-16-04-ee64c073cd79)

# STEP BY STEP INSTALLATION GUIDE
Note: The following guide accommodates many different Linux distros/installations. You may see error messages, primarily with yum based installation. In particular, the default method tries to use snap to install go platform but we install it directly with yum method.
## Update System and Install Essential Tools
### Ubuntu/Debian Based e.g. Ubuntu, Mint, debian (using apt package manager)
        sudo apt update
        sudo apt-get install build-essential
        sudo apt install git

### RHEL Based e.g. Centos, RedHat, Fedora (using yum package manager)
Assume superuser privilege (alternatively, prefix necessary commands with sudo e.g. sudo yum update -y)

        su
        
Install Essential Tools (Note some of these will already be installed depending on distribution)

        yum update -y
        yum install -y openssl-devel
        yum groupinstall -y "Development Tools"
        yum install -y git
        yum install -y wget
        yum install -y make
        yum install -t g++
        
Install go platform

        wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.13.linux-amd64.tar.gz
        rm go1.13.linux-amd64.tar.gz
        
Set up paths, this needs to be done each session. Alternatively, append paths in $HOME/.bash_profile

        export PATH=$PATH:/usr/local/go/bin
        export LD_LIBRARY_PATH=/usr/local/lib
                
## Set Up directories and clone gosdk
### (Common) starting in $HOME folder
        mkdir go
        cd go
        mkdir github.com
        cd github.com
        mkdir 0chain
        cd 0chain
        git clone https://github.com/0chain/gosdk.git

## Build SDK/Tools
### (Common) starting in 0chain folder
        cd gosdk
        make build-tools
        make install
        make install-herumi
        make install-gosdk
        make clean
        cd ..

## Test SDK (Optional)

### (Common) starting in 0chain folder
        cd gosdk/_sdkver
        go build -o sdkver sdkver.go
        cd ../..
        gosdk/_sdkver/sdkver

(Should output SDK version if successful)

## Build CLI Tools
### (Common) starting in 0chain folder

ZBox:-

        git clone https://github.com/0chain/zboxcli.git
        cd zboxcli
        go build -tags bn256 -o zbox

To test:-
        
        ./zbox
        
Then back to 0chain folder:-

        cd ..

Zwallet

        git clone https://github.com/0chain/zwalletcli.git
        cd zwalletcli
        go build -tags bn256 -o zwallet

You will at least require $HOME/.zcn folder with nodes.yaml file present to test wallet (change this as required)

        mkdir $HOME/.zcn
        cp sample/config/devb.yml $HOME/.zcn/nodes.yaml

To test:-

        ./zwallet
        
Then back to 0chain folder:-

        cd ..

## DEPLOYMENT GUIDE
### (Based on having already built CLI Tools as above)

- Target Machine must be recent kernel as distributions mentioned, (specifically libc6 version >= 2.27?)

NOTE: Do not attempt to install libc version into older Linux installation (unless you know what you are doing) as it can mess up your system

Supported platforms all of the above e.g.
- RHEL7+, Centos7+, Ubuntu18+, Mint19+, Fedora30+ (and derivatives)

PLUS
- Debian 10+

### Copy files and configure

Copy these files From/To /usr/local/lib

        libmcl.so
        libbls256.so

Perform this command on target machine for each session required or add path in $HOME/.bash_profile

        export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib

Adjust permissions for user as needed.

Copy other files as required, e.g.

        zbox
        zwallet
        nodes.yaml

into folder

        $HOME/.zcn
