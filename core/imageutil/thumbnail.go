package imageutil

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/0chain/gosdk/core/logger"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var (
	//go:embed image_rs/image_rs.wasm
	imageWasm []byte

	imageRs    *ImageRs
	converters []Converter
	logging    *logger.Logger
)

func init() {
	var err error
	imageRs, err = NewImageRs()
	if err != nil {
		panic(err)
	}
	converters = []Converter{imageRs}
	logging = &logger.Logger{}
	logging.Init(3, "imageutil")
}

type Option struct {
	// Format of input image
	IFormat string `json:"input_format,omitempty"`
}

func Thumbnail(img []byte, width, height int, options string) (ConvertRes, error) {
	var opt Option
	err := json.Unmarshal([]byte(options), &opt)
	if err != nil {
		return ConvertRes{}, err
	}
	for _, converter := range converters {
		if !converter.IsFormatSupported(opt.IFormat) {
			continue
		}
		res, err := converter.Convert(img, width, height, ConvertOptions{})
		if err == nil {
			return res, nil
		}
		logging.Error("convertor failed to convert", err)
	}
	for _, converter := range converters {
		res, err := converter.Convert(img, width, height, ConvertOptions{})
		if err == nil {
			return res, nil
		}
		logging.Error(err)
	}
	return ConvertRes{}, errors.New("all converters failed to convert; use default thumbnail image")
}

type ConvertOptions struct{}

type ConvertRes struct {
	// thumbnail image
	ThumbnailImg []byte `json:"thumbnail_img,omitempty"`
	// format of thumbnail image
	Format string	`json:"format,omitempty"`
}

type Converter interface {
	Convert([]byte, int, int, ConvertOptions) (ConvertRes, error)
	IsFormatSupported(format string) bool
}

type ImageRs struct {
	supportedFormats map[string]bool
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
	supportedFormats := map[string]bool{
		"bmp": true, "dds": true, "exr": true, "ff": true, "gif": true,
		"hdr": true, "ico": true, "jpeg": true, "png": true, "pnm": true, "qoi": true,
		"tga": true, "tiff": true, "webp": true,
	}
	return &ImageRs{
		supportedFormats: supportedFormats,
		ctx:              ctx,
		runtime:          runtime,
		compiledMod:      compiledMod,
	}, nil
}

func (i *ImageRs) Convert(img []byte, width, height int, co ConvertOptions) (ConvertRes, error) {
	var errW bytes.Buffer
	mod, err := i.runtime.InstantiateModule(i.ctx, i.compiledMod, wazero.NewModuleConfig().WithStderr(&errW))
	if err != nil {
		return ConvertRes{}, fmt.Errorf("failed to instantiate module: %v", err)
	}

	allocate := mod.ExportedFunction("allocate")
	deallocate := mod.ExportedFunction("deallocate")
	thumbnail := mod.ExportedFunction("thumbnail")

	imgLen := len(img)
	results, err := allocate.Call(i.ctx, uint64(imgLen))
	if err != nil {
		return ConvertRes{}, fmt.Errorf("error allocating memory: %v", err)
	}
	ptr := results[0]
	defer func() {
		_, err = deallocate.Call(i.ctx, ptr, uint64(imgLen))
		if err != nil {
			logging.Error("error deallocating memory: ", err)
		}
	}()

	if !mod.Memory().Write(uint32(ptr), img) {
		return ConvertRes{}, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d",
			ptr, imgLen, mod.Memory().Size())
	}

	ptrSize, err := thumbnail.Call(i.ctx, ptr, uint64(imgLen), uint64(width), uint64(height))
	if err != nil {
		return ConvertRes{}, fmt.Errorf("err calling thumbnail: %v", err)
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
		return ConvertRes{}, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
			thumbnailPtr, thumbnailSize, mod.Memory().Size())
	}

	if len(res) == 0 {
		return ConvertRes{}, fmt.Errorf("error occurred : %v", errW.String())
	}

	cr := ConvertRes{}
	cr.ThumbnailImg = append(cr.ThumbnailImg, res...)
	cr.Format = "jpeg"
	return cr, nil
}

func (i *ImageRs) IsFormatSupported(format string) bool {
	return i.supportedFormats[format]
}
