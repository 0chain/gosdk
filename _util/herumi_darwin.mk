ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

ifneq ($(PLATFORM), darwin)
$(error PLATFORM doesn't match darwin)
endif

.PHONY: install-herumi-deps install-gmp install-openssl upgrade-gmp upgrade-openssl

install-herumi-deps: install-openssl install-gmp

install-openssl:
	@echo "Installing openssl ..."
	brew install openssl
	$(shell sudo ln -sf /usr/local/opt/openssl/lib/libcrypto.dylib /usr/local/lib/)

upgrade-openssl:
	@echo "Upgrading openssl ..."
	brew upgrade openssl
	$(shell sudo ln -sf /usr/local/opt/openssl/lib/libcrypto.dylib /usr/local/lib/)

install-gmp:
	@echo "Installing gmp ..."
	brew install gmp
	$(shell sudo ln -sf /usr/local/Cellar/gmp/*/lib /usr/local/lib)

upgrade-gmp:
	@echo "Upgrading gmp ..."
	brew upgrade gmp
	$(shell sudo ln -sf /usr/local/Cellar/gmp/*/lib /usr/local/lib)

