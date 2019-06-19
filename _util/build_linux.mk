ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), linux)
$(error PLATFORM doesn't match linux)
endif

build-tools:
	@echo ">>> Update apt"
	sudo apt update
	@echo ">>> Install jq"
	sudo apt-get install jq
	@echo ">>> Installing build-essentials tools"
	sudo apt-get -y install build-essential
	@echo ">>> Install go tools"
	sudo snap install go --classic
	@echo ">>> Display go version"
	go version
