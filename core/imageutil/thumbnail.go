package imageutil

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/0chain/gosdk/core/logger"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var (
	//go:embed image_rs/image_rs.wasm
	imageWasm []byte

	imageRs *ImageRs
	logging *logger.Logger
)

const (
	DefWidth  = 120
	DefHeight = 90
)

func init() {
	var err error
	imageRs, err = NewImageRs()
	if err != nil {
		panic(err)
	}
	logging = &logger.Logger{}
	logging.Init(4, "imageutil")
}

func Thumbnail(img []byte, width, height int) ([]byte, error) {
	return imageRs.Convert(img, width, height)
}

type ImageRs struct {
	ctx         context.Context
	runtime     wazero.Runtime
	compiledMod wazero.CompiledModule
	apiMod      api.Module
}

func NewImageRs() (*ImageRs, error) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	compiledMod, err := runtime.CompileModule(ctx, imageWasm)
	if err != nil {
		return nil, fmt.Errorf("error compiling imageWasm: %v", err)
	}
	apiMod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %v", err)
	}
	return &ImageRs{
		ctx:         ctx,
		runtime:     runtime,
		compiledMod: compiledMod,
		apiMod:      apiMod,
	}, nil
}

func (i *ImageRs) Convert(img []byte, width, height int) ([]byte, error) {
	allocate := i.apiMod.ExportedFunction("allocate")
	deallocate := i.apiMod.ExportedFunction("deallocate")
	thumbnail := i.apiMod.ExportedFunction("thumbnail")

	imgLen := len(img)
	results, err := allocate.Call(i.ctx, uint64(imgLen))
	if err != nil {
		return nil, fmt.Errorf("error allocating memory: %v", err)
	}
	ptr := results[0]
	defer func() {
		_, err = deallocate.Call(i.ctx, ptr, uint64(imgLen))
		if err != nil {
			logging.Error("error deallocating memory: ", err)
		}
	}()

	if !i.apiMod.Memory().Write(uint32(ptr), img) {
		return nil, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d",
			ptr, imgLen, i.apiMod.Memory().Size())
	}

	ptrSize, err := thumbnail.Call(i.ctx, ptr, uint64(imgLen), uint64(width), uint64(height))
	if err != nil {
		return nil, fmt.Errorf("err calling thumbnail: %v", err)
	}
	thumbnailPtr := uint32(ptrSize[0] >> 32)
	thumbnailSize := uint32(ptrSize[0])
	defer func() {
		_, err = deallocate.Call(i.ctx, uint64(thumbnailPtr), uint64(thumbnailSize))
		if err != nil {
			logging.Error("error deallocating thumbnailPtr: ", err)
		}
	}()

	res, ok := i.apiMod.Memory().Read(thumbnailPtr, thumbnailSize)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
			thumbnailPtr, thumbnailSize, i.apiMod.Memory().Size())
	}

	return res, nil
}
