package imageutil

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"

	"github.com/0chain/gosdk/core/logger"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var (
	//go:embed image_rs/image_rs.wasm
	imageWasm []byte

	imageRs    *ImageRs
	logging    *logger.Logger
)

const (
	DefWidth = 120
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
	ctx              context.Context
	runtime          wazero.Runtime
	compiledMod      wazero.CompiledModule
}

func NewImageRs() (*ImageRs, error) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	compiledMod, err := runtime.CompileModule(ctx, imageWasm)
	if err != nil {
		return nil, fmt.Errorf("error compiling imageWasm: %v", err)
	}
	return &ImageRs{
		ctx:              ctx,
		runtime:          runtime,
		compiledMod:      compiledMod,
	}, nil
}

func (i *ImageRs) Convert(img []byte, width, height int) ([]byte, error) {
	var errW bytes.Buffer
	mod, err := i.runtime.InstantiateModule(i.ctx, i.compiledMod, wazero.NewModuleConfig().WithStderr(&errW))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %v", err)
	}

	allocate := mod.ExportedFunction("allocate")
	deallocate := mod.ExportedFunction("deallocate")
	thumbnail := mod.ExportedFunction("thumbnail")

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

	if !mod.Memory().Write(uint32(ptr), img) {
		return nil, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d",
			ptr, imgLen, mod.Memory().Size())
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

	res, ok := mod.Memory().Read(thumbnailPtr, thumbnailSize)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
			thumbnailPtr, thumbnailSize, mod.Memory().Size())
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("error occurred : %v", errW.String())
	}

	return res, nil
}
