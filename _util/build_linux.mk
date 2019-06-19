ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), linux)
$(error PLATFORM doesn't match linux)
endif

build-tools:
	@echo "Installing build-essentials tools"
	sudo apt-get -y install build-essential
	@echo "Install go tools"
	sudo snap -y install go --classic
