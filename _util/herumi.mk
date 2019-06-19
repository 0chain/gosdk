ifndef PLATFORM
$(error PLATFORM is not set. Unable to add platform specific targets)
endif

HERUMI_TARGETS:=\
	clone-herumi- \
	checkout-herumi \
	build-herumi \
	test-herumi \
	install-herumi \
	help

#Master having build issues
$(HERUMI_TARGETS): mcl_branch=master
$(HERUMI_TARGETS): bls_branch=master

HERUMI_DIR?=$(ROOT_DIR)/_herumi
BLS_DIR?=$(HERUMI_DIR)/bls
MCL_DIR?=$(HERUMI_DIR)/mcl
NPROC:=8

include _util/herumi_$(PLATFORM).mk

install-herumi-deps: install-openssl install-gmp

.PHONY: build-tools herumi-clone herumi-build herumi-install

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

PREFIX?=/usr/local

install-herumi: install-herumi-deps | clone-herumi checkout-herumi build-herumi test-herumi
	@$(PRINT_MAG)
	@echo "Installing MCL: branch=$(mcl_branch)"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(MCL_DIR) install
	@$(PRINT_MAG)
	@echo "Installing BLS: branch=$(bls_branch)"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(BLS_DIR) install
	@sudo $(MAKE) ldload-herumi
