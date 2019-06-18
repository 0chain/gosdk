ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), linux)
$(error PLATFORM doesn't match linux)
endif

.PHONY: install-herumi-deps install-gmp install-openssl upgrade-gmp upgrade-openssl

install-herumi-deps: install-openssl install-gmp

install-openssl:
	@echo "Installing openssl ..."
	sudo apt-get -y install libssl-dev

upgrade-openssl:
	@echo "Upgrading openssl ..."
	sudo apt-get -y install libssl-dev

install-gmp:
	@echo "Installing gmp ..."
	sudo apt-get -y install libgmp3-dev

upgrade-gmp:
	@echo "Upgrading gmp ..."
	sudo apt-get -y upgrade libgmp3-dev

ldload-herumi:
	@echo "Loading herumi library - linux"
	@sudo ldconfig
	@sudo ldconfig -p | egrep "bls|mcl"
