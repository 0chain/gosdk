0CHAIN_PATH	:=  github.com/0chain
GOSDK_PATH :=  $(0CHAIN_PATH)/gosdk
OUTDIR := $(ROOT_DIR)/out
IOSMOBILESDKDIR     := $(OUTDIR)/iossdk
ANDROIDMOBILESDKDIR := $(OUTDIR)/androidsdk
IOSBINNAME 		:= zcncore.xcframework
ANDROIDBINNAME	:= zcncore.aar

.PHONY: build-mobilesdk setup-gomobile build-iossimulator build-ios

$(IOSMOBILESDKDIR):
	$(shell mkdir -p $(IOSMOBILESDKDIR)/lib)

$(ANDROIDMOBILESDKDIR):
	$(shell mkdir -p $(ANDROIDMOBILESDKDIR)/lib)

setup-gomobile: $(IOSMOBILESDKDIR) $(ANDROIDMOBILESDKDIR)
	@$(PRINT_MAG)
	@echo "============================================================"
	@echo "    Initializing gomobile. Please wait it may take a while ..."
	@echo "------------------------------------------------------------"
	@go get -d golang.org/x/mobile/cmd/gomobile
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

build-mobilesdk: $(GOPATH)/src/$(GOSDK_PATH)
ifeq ($(filter-out undefined,$(foreach v, IOS ANDROID,$(origin $(v)))),)
	@$(PRINT_RED)
	@echo ""
	@echo "Usage:"
	@echo '   For iOS and Android: make build-mobilesdk IOS=1 ANDROID=1'
	@echo '   For iOS only: make build-mobilesdk IOS=1'
	@echo '   For Android only: make build-mobilesdk ANDROID=1'
endif
	@$(PRINT_CYN)
ifneq ($(IOS),)
	@echo "Building iOS framework. Please wait..."
	@gomobile bind -ldflags="-s -w" -target=ios -tags mobile -o $(IOSMOBILESDKDIR)/$(IOSBINNAME) $(GOSDK_PATH)/zcncore
	@echo "   $(IOSMOBILESDKDIR)/$(IOSBINNAME). - [OK]"
endif
ifneq ($(ANDROID),)
	@echo "Building Android framework. Please wait..."
	@gomobile bind -target=android/arm64,android/amd64 -tags mobile -ldflags=-extldflags=-Wl,-soname,libgojni.so -o $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME) $(GOSDK_PATH)/zcncore $(GOSDK_PATH)/core/common
	@echo "   $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME). - [OK]"
endif
	@echo ""
	@$(PRINT_NON)

clean-mobilesdk:
	@rm -rf $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME)
	@rm -rf $(IOSMOBILESDKDIR)/$(IOSBINNAME)

cleanall-gomobile:
	@rm -rf $(OUTDIR)
	@rm -rf $(BLS_LIB_BASE_PATH)

gomobile-install:
	go install golang.org/x/mobile/cmd/gomobile@latest
	gomobile init

build-iossimulator: 
	@echo "Building iOS Simulator framework. Please wait..."
	@@gomobile bind -v -ldflags="-s -w" -target=iossimulator -tags "ios iossimulator mobile" -o $(IOSMOBILESDKDIR)/simulator/$(IOSBINNAME) $(GOSDK_PATH)/zcncore $(GOSDK_PATH)/core/common $(GOSDK_PATH)/mobilesdk/sdk
	@echo "   $(IOSMOBILESDKDIR)/simulator/$(IOSBINNAME). - [OK]"

build-ios: 
	@echo "Building iOS framework. Please wait..."
	@@gomobile bind -v -ldflags="-s -w" -target=ios/arm64,iossimulator/amd64 -tags "ios mobile" -o $(IOSMOBILESDKDIR)/ios/$(IOSBINNAME) $(GOSDK_PATH)/zcncore $(GOSDK_PATH)/core/common $(GOSDK_PATH)/mobilesdk/sdk
	@echo "   $(IOSMOBILESDKDIR)/ios/$(IOSBINNAME). - [OK]"	

build-android: 
	@echo "Building Android framework. Please wait..."
	@gomobile bind -target=android/arm64 -tags mobile -ldflags=-extldflags=-Wl,-soname,libgojni.so -o $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME) $(GOSDK_PATH)/zcncore $(GOSDK_PATH)/core/common $(GOSDK_PATH)/mobilesdk/sdk
	@echo "   $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME). - [OK]"