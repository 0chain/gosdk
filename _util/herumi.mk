HERUMI_TARGETS:=herumi-clone herumi-checkout herumi-build herumi-install show

#Master having build issues
#$(HERUMI_TARGETS): mcl_branch=master
#$(HERUMI_TARGETS): bls_branch=master

#Success Branch:
$(HERUMI_TARGETS): bls_branch?=b1733a744a2e53a828806b121c4a5cb681c5f94b
$(HERUMI_TARGETS): mcl_branch?=master

HERUMI_DIR?=$(ROOT_DIR)/_herumi
BLS_DIR?=$(HERUMI_DIR)/bls
MCL_DIR?=$(HERUMI_DIR)/mcl
NPROC:=8

herumi-deps: openssl-install gmp-install

openssl-install:
	@echo "Installing openssl ..."
	brew install openssl
	$(shell sudo ln -sf /usr/local/opt/openssl/lib/libcrypto.dylib /usr/local/lib/)

openssl-upgrade:
	@echo "Upgrading openssl ..."
	brew upgrade openssl
	$(shell sudo ln -sf /usr/local/opt/openssl/lib/libcrypto.dylib /usr/local/lib/)

gmp-install:
	@echo "Installing gmp ..."
	brew install gmp
	$(shell sudo ln -sf /usr/local/Cellar/gmp/*/lib /usr/local/lib)

gmp-upgrade:
	@echo "Upgrading gmp ..."
	brew upgrade gmp
	$(shell sudo ln -sf /usr/local/Cellar/gmp/*/lib /usr/local/lib)

.PHONY: herumi-clone herumi-build herumi-install

herumi-clone:
	@echo Deleting directories: [$(BLS_DIR) $(MCL_DIR)]
	@rm -rf $(BLS_DIR) $(MCL_DIR)
	git clone http://github.com/herumi/mcl.git $(MCL_DIR)
	git clone http://github.com/herumi/bls.git $(BLS_DIR)

herumi-checkout:
	@echo Checking out BLS: branch=$(bls_branch)
	cd $(BLS_DIR); git checkout $(bls_branch)
	@echo Checking out MCL: branch=$(mcl_branch)
	cd $(MCL_DIR); git checkout $(mcl_branch)

herumi-build:
	@$(PRINT_MAG)
	@echo "Building BLS: branch=$(bls_branch)"
	@$(PRINT_NON)
	$(MAKE) -C $(BLS_DIR) -j $(NPROC) lib/libbls256.a
	@$(PRINT_MAG)
	@echo "Building MCL: branch=$(mcl_branch)"
	@$(PRINT_NON)
	$(MAKE) -C $(MCL_DIR) -j $(NPROC) lib/libmclbn256.a

herumi-install:
	@$(PRINT_MAG)
	@echo "Installing BLS: branch=$(bls_branch)"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(MCL_DIR) install
	@$(PRINT_MAG)
	@echo "Installing BLS: branch=$(bls_branch)"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(BLS_DIR) install

herumi-all: | herumi-clone herumi-checkout herumi-deps herumi-build herumi-install
