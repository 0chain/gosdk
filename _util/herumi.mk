HERUMI_TARGETS:=\
	clone-herumi- \
	checkout-herumi \
	build-herumi \
	test-herumi \
	install-herumi-official \
	install-herumi-workaround \
	help

#Master having build issues
$(HERUMI_TARGETS): mcl_branch=master
$(HERUMI_TARGETS): bls_branch=master

#Success Branch:
#$(HERUMI_TARGETS): bls_branch?=b1733a744a2e53a828806b121c4a5cb681c5f94b
#$(HERUMI_TARGETS): mcl_branch?=master

HERUMI_DIR?=$(ROOT_DIR)/_herumi
BLS_DIR?=$(HERUMI_DIR)/bls
MCL_DIR?=$(HERUMI_DIR)/mcl
NPROC:=8

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

.PHONY: herumi-clone herumi-build herumi-install

clone-herumi:
	@echo Deleting directories: [$(BLS_DIR) $(MCL_DIR)]
	@rm -rf $(BLS_DIR) $(MCL_DIR)
	git clone http://github.com/herumi/mcl.git $(MCL_DIR)
	git clone http://github.com/herumi/bls.git $(BLS_DIR)

checkout-herumi:
	@echo Checking out BLS: branch=$(bls_branch)
	cd $(BLS_DIR); git checkout $(bls_branch)
	@echo Checking out MCL: branch=$(mcl_branch)
	cd $(MCL_DIR); git checkout $(mcl_branch)

build-herumi:
	@$(PRINT_MAG)
	@echo "Building BLS: branch=$(bls_branch)"
	@$(PRINT_NON)
	$(MAKE) -C $(BLS_DIR) -j $(NPROC) lib/libbls256.a
	@$(PRINT_MAG)
	@echo "Building MCL: branch=$(mcl_branch)"
	@$(PRINT_NON)
	$(MAKE) -C $(MCL_DIR) -j $(NPROC) lib/libmclbn256.a

test-herumi:
	$(MAKE) -C $(BLS_DIR) test_go256

clean-herumi:
	@rm -rf $(HERUMI_DIR)/

install-herumi-official:
	@$(PRINT_MAG)
	@echo "Installing MCL: branch=$(mcl_branch)"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(MCL_DIR) install
	@$(PRINT_MAG)
	@echo "Installing BLS: branch=$(bls_branch)"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(BLS_DIR) install

PREFIX?=/usr/local

install-herumi-workaround:
	@$(PRINT_MAG)
	@echo "Workaround MCL:: branch=$(mcl_branch)"
	@$(PRINT_NON)
	sudo cp -a $(MCL_DIR)/include/mcl/ $(PREFIX)/include/mcl
	@$(PRINT_MAG)
	@echo "Workaround BLS: branch=$(bls_branch)"
	@$(PRINT_NON)
	sudo cp -a $(BLS_DIR)/include/bls/ $(PREFIX)/include/bls

install-herumi: install-herumi-deps | \
	clone-herumi \
	checkout-herumi \
	build-herumi \
	test-herumi \
	install-herumi-official \
	install-herumi-workaround

