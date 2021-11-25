ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), linux)
$(error PLATFORM doesn't match linux)
endif

REDHATOS := $(shell command cat /etc/redhat-release 2> /dev/null)

build-tools:
ifdef REDHATOS
	@echo ">>> Update apt"
	yum update
	@echo ">>> Install jq"
	yum install -y epel-release
	yum install -y jq
	@echo ">>> Installing build-essentials tools"
	yum install -y gcc gcc-c++ kernel-devel make
	yum groupinstall -y "Development Tools"
	@echo ">>> Install go tools"
	yum install -y wget
	wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz
	tar -C /usr/local -xzf go1.13.linux-amd64.tar.gz
	@echo ">>> ADD GO BIN PATH to path manually, then test go"
	@echo "export PATH=$PATH:/usr/local/go/bin"
	@echo "go version"
else
	@echo ">>> Update apt"
	sudo apt -y update
	@echo ">>> Install jq"
	sudo apt-get install -y jq
	@echo ">>> Installing build-essentials tools"
	sudo apt-get -y install build-essential
	@echo ">>> Install go tools"
	sudo snap install go --classic
	@echo ">>> Display go version"
	go version
endif
