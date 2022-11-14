0CHAIN_PATH	:=  github.com/0chain
GOSDK_PATH :=  $(0CHAIN_PATH)/gosdk
OUTDIR := $(ROOT_DIR)/out
IOSMOBILESDKDIR     := $(OUTDIR)/iossdk
ANDROIDMOBILESDKDIR := $(OUTDIR)/androidsdk
MACSDKDIR	:= $(OUTDIR)/macossdk
IOSBINNAME 		:= zcncore.xcframework
ANDROIDBINNAME	:= zcncore.aar

PKG_EXPORTS := $(GOSDK_PATH)/zcncore $(GOSDK_PATH)/core/common $(GOSDK_PATH)/mobilesdk/sdk $(GOSDK_PATH)/mobilesdk/zbox

.PHONY: setup-gomobile build-iossimulator build-ios build-android build-android-debug

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
	@@gomobile bind -v -ldflags="-s -w" -target=iossimulator -tags "ios iossimulator mobile" -o $(IOSMOBILESDKDIR)/simulator/$(IOSBINNAME) $(PKG_EXPORTS)
	@echo "   $(IOSMOBILESDKDIR)/simulator/$(IOSBINNAME). - [OK]"

build-ios: 
	@echo "Building iOS framework. Please wait..."
	@@gomobile bind -v -ldflags="-s -w" -target=ios/arm64,iossimulator/amd64 -tags "ios mobile" -o $(IOSMOBILESDKDIR)/ios/$(IOSBINNAME) $(PKG_EXPORTS)
	@echo "   $(IOSMOBILESDKDIR)/ios/$(IOSBINNAME). - [OK]"	

build-android: 
	@echo "Building Android framework. Please wait..."
	@gomobile bind -v -ldflags="-s -w -extldflags=-Wl,-soname,libgojni.so" -target=android/arm64,android/amd64 -tags mobile  -o $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME) $(PKG_EXPORTS)
	@echo "   $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME). - [OK]"

build-android-debug: 
	@echo "Building Android framework. Please wait..."
	@gomobile bind -v -ldflags="-extldflags=-Wl" -gcflags '-N -l' -target=android/arm64,android/amd64 -tags mobile  -o $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME) $(PKG_EXPORTS)
	@echo "   $(ANDROIDMOBILESDKDIR)/$(ANDROIDBINNAME). - [OK]"

build-macos: 
	@echo "Building MAC framework. Please wait..."
	@gomobile bind -v -ldflags="-s -w" -target=macos -tags mobile -o $(MACSDKDIR)/$(IOSBINNAME) $(PKG_EXPORTS)
	@echo "   $(MACSDKDIR)/$(IOSBINNAME). - [OK]"