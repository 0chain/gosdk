# 0chain/gosdk
The 0chain client SDK is written in GO. 
This repository currently supports the following platforms:
- OSX Mojave 10.14.5 
- LINUX (Ubuntu/bionic). 

It is possible to support the sdk for other variations of Linux as well. 

## Build and Installation 
0chain/gosdk is build and installed using [GNU Make](https://www.gnu.org/software/make/). 
The Makefile has following targets to ease installation of build tools and the 0chain/gosdk. 

**Success of installation of the library and GO modules is highly dependent upon the prior installed
packages on the developer system.**

These steps have been tested out thoroughly on OSX Mojave 10.14.5 and Vanilla Ubuntu BIONIC. 

0chain/gosdk has heavy dependency on [BLS](https://github.com/herumi/bls) and [MCL](https://github.com/herumi/mcl) 
provided by [MITSUNARI Shigeo](https://github.com/herumi). Developers should refer to those links when they encounter any errors. 

Please send email to [alphanet@chain.net](mailto:alphanet@0chain.net) if you encounter any problems.

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

