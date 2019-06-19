ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), darwin)
$(error PLATFORM doesn't match darwin)
endif

build-tools:
	@echo ">>> Install go"
	brew install go
	@echo ">>> Install jq"
	brew install jq

