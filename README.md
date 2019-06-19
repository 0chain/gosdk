# 0chain - gosdk
Client SDK Written in GO. The gosdk is currently supported for OSX and LINUX (Ubuntu/bionic). 
The sdk is written in GO. The SDK has dependency on BLS and MCL provided by MITSUNARI Shigeo. 
Please refer to the following links and repositories for more information.

- [MITSUNARI Shigeo](https://github.com/herumi)
- [BLS](https://github.com/herumi/bls)
- [MCL](https://github.com/herumi/mcl)


## Pre-requisites
The Makefile has following targets to ease installation of build tools and the GOSDK. 
Success of installation of the library and GO modules is highly dependent upon the developer environment. 
These steps have been tested out thoroughly on OSX Mojave 10.14.5 and Ubuntu BIONIC. 

Please send email to [partners](mailto:partners@0chain.net) if you encounter any problems.

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

- How to install GO on ubuntu using snap
```.env
sudo snap install go --classic
```
- [How to: Install Go 1.11.2 on Ubuntu](https://medium.com/@patdhlk/how-to-install-go-1-9-1-on-ubuntu-16-04-ee64c073cd79)

