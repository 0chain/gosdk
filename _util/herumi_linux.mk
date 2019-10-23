ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), linux)
$(error PLATFORM doesn't match linux)
endif

REDHATOS := $(shell command cat /etc/redhat-release 2> /dev/null)

.PHONY: install-herumi-deps install-gmp install-openssl upgrade-gmp upgrade-openssl

install-herumi-deps: install-openssl install-gmp

install-openssl:
	@echo "Installing openssl ..."
ifdef REDHATOS
	yum install openssl-devel
else
	sudo apt-get -y install libssl-dev
endif

upgrade-openssl:
	@echo "Upgrading openssl ..."
ifdef REDHATOS
	yum update openssl-devel
else
	sudo apt-get -y install libssl-dev
endif

install-gmp:
	@echo "Installing gmp ..."
ifdef REDHATOS
	yum install gmp-devel
else
	sudo apt-get -y install libgmp3-dev
endif

upgrade-gmp:
	@echo "Upgrading gmp ..."
ifdef REDHATOS
	yum update gmp-devel
else
	sudo apt-get -y upgrade libgmp3-dev
endif

ldload-herumi:
ifdef REDHATOS
	@echo ">>> SET LD_LIBRARY_PATH manually as follows"
	@echo "export LD_LIBRARY_PATH=/usr/local/lib"
else
	@echo "Loading herumi library - linux"
	@sudo ldconfig
	@sudo ldconfig -p | egrep "bls|mcl"
endif
