0CHAIN_PATH	:=  github.com/0chain
GOSDK_PATH :=  $(0CHAIN_PATH)/gosdk
OUTDIR := $(ROOT_DIR)/out
IOSMOBILESDKDIR     := $(OUTDIR)/0chainiosmobilesdk
ANDROIDMOBILESDKDIR := $(OUTDIR)/0chainandroidmobilesdk
IOSBINNAME 		:= zcncore.framework
ANDROIDBINNAME	:= zcncore.aar

.PHONY: build-mobilesdk

BLS_LIB_BASE_PATH=$(GOPATH)/src/github.com/herumi
export CGO_CFLAGS+=-I$(BLS_LIB_BASE_PATH)/bls/include -I$(BLS_LIB_BASE_PATH)/mcl/include

$(BLS_LIB_BASE_PATH):
	@git clone http://github.com/herumi/mcl.git $(BLS_LIB_BASE_PATH)/mcl
	@cd $(BLS_LIB_BASE_PATH)/mcl && git checkout cc9762f14f7f6d4bbc29c0ca418781af4a74f92d && cd - >/dev/null
	@git clone http://github.com/herumi/bls.git $(BLS_LIB_BASE_PATH)/bls
	@cd $(BLS_LIB_BASE_PATH)/bls && git checkout 058c89ea4262b6131e704f12583b2a852462d4f9 && cd - >>/dev/null
	$(eval NCPU=$(shell sysctl -n hw.ncpu))
	@$(PRINT_MAG)
	@echo "============================================================"
	@echo "    Building BLS for MAC...                       "
	@echo "------------------------------------------------------------"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(BLS_LIB_BASE_PATH)/bls -j$(NCPU) install
	@$(PRINT_MAG)
	@echo "============================================================"
	@echo "    Building MCL for MAC...                       "
	@echo "------------------------------------------------------------"
	@$(PRINT_NON)
	@sudo $(MAKE) -C $(BLS_LIB_BASE_PATH)/mcl -j$(NCPU) lib/libmclbn256.dylib install

$(IOSMOBILESDKDIR):
	$(shell mkdir -p $(IOSMOBILESDKDIR)/lib)

$(ANDROIDMOBILESDKDIR):
	$(shell mkdir -p $(ANDROIDMOBILESDKDIR)/lib)

setup-gomobile: $(BLS_LIB_BASE_PATH) $(IOSMOBILESDKDIR) $(ANDROIDMOBILESDKDIR)
	@cd $(BLS_LIB_BASE_PATH)/bls && git checkout . && git checkout 058c89ea && git apply $(ROOT_DIR)/patches/github.com-herumi-bls-gomobile_ios.patch && cd -
	@$(PRINT_CYN)
	@echo "============================================================"
	@echo "    Building BLS for iOS                    "
	@echo "------------------------------------------------------------"
	@$(PRINT_NON)
	@$(MAKE) -C $(BLS_LIB_BASE_PATH)/bls gomobile_ios CURVE_BIT=256
	@$(MAKE) -C $(BLS_LIB_BASE_PATH)/bls gomobile_ios CURVE_BIT=384
	@cp -Rf $(BLS_LIB_BASE_PATH)/bls/ios/* $(IOSMOBILESDKDIR)/lib
ifeq ($(NOANDROID),)
	@$(PRINT_CYN)
	@echo "============================================================"
	@echo "    Building BLS for Android                    "
	@echo "------------------------------------------------------------"
	@$(PRINT_NON)
	@$(MAKE) -C $(BLS_LIB_BASE_PATH)/bls gomobile_android CURVE_BIT=256
	@$(MAKE) -C $(BLS_LIB_BASE_PATH)/bls gomobile_android CURVE_BIT=384
	@cp -Rf $(BLS_LIB_BASE_PATH)/bls/android/* $(ANDROIDMOBILESDKDIR)/lib
endif
	@$(PRINT_MAG)
	@echo "============================================================"
	@echo "    Initializing gomobile. Please wait it may take a while ..."
	@echo "------------------------------------------------------------"
	@go get golang.org/x/mobile/cmd/gomobile
	@$(PRINT_NON)
	@gomobile init
	@$(PRINT_GRN)
	@echo "  ___ __  _  _ ____ __   ____ ____ ____ ____  "
	@echo " / __/  \( \/ (  _ (  ) (  __(_  _(  __(    \ "
	@echo "( (_(  O / \/ \) __/ (_/\) _)  )(  ) _) ) D ( "
	@echo " \___\__/\_)(_(__) \____(____)(__)(____(____/ "
	@$(PRINT_NON)

$(GOPATH)/src/$(GOSDK_PATH):
	@echo "gosdk is not in GOPATH. Creating softlink..."
ifneq ($(GOPATH), )
	$(shell ln -sf $(ROOT_DIR) $(GOPATH)/src/$(0CHAIN_PATH))
endif

build-mobilesdk: $(GOPATH)/src/$(GOSDK_PATH) getrev
	@$(PRINT_CYN)
	@echo "Building iOS framework. Please wait..."
	@cd $(BLS_LIB_BASE_PATH)/bls && git checkout . && git apply $(ROOT_DIR)/patches/github.com-herumi-bls-gomobile_ios.patch && cd - >> /dev/null
	@gomobile bind -ldflags="-s -w" -target=ios -o $(IOSMOBILESDKDIR)/$(IOSBINNAME) $(GOSDK_PATH)/zcncore
	@echo "   $(IOSMOBILESDKDIR)/$(IOSBINNAME). - [OK]"
ifeq ($(NOANDROID),)
	@echo "Building Android framework. Please wait..."
	@cd $(BLS_LIB_BASE_PATH)/bls && git checkout . && git apply $(ROOT_DIR)/patches/github.com-herumi-bls-gomobile_android.patch && cd - >> /dev/null
	@gomobile bind -target=android/arm64,android/amd64 -ldflags=-extldflags=-Wl,-soname,libgojni.so -o $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME) $(GOSDK_PATH)/zcncore
	@echo "   $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME). - [OK]"
endif
	@cd $(BLS_LIB_BASE_PATH)/bls && git checkout . && cd - >> /dev/null
	@$(PRINT_NON)

clean-mobilesdk:
	@rm -rf $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME)
	@rm -rf $(IOSMOBILESDKDIR)/$(IOSBINNAME)

cleanall-gomobile:
	@rm -rf $(OUTDIR)
	@rm -rf $(BLS_LIB_BASE_PATH)