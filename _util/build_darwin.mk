ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), darwin)
$(error PLATFORM doesn't match darwin)
endif

build-tools:
	brew install go
	brew install jq

