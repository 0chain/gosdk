package sdk

import (
	"path/filepath"
	"strconv"
	"strings"
)

// FileNameBuilder build file name based output format
type FileNameBuilder interface {
	OutDir() string
	FileExt() string
	OutFile() string
	ClipsFile(index int) string
	ClipsFileName(index int) string
}

var (
	builderFactory = make(map[string]func(outDir, fileName, fileExt string) FileNameBuilder)
)

func init() {
	builderFactory[".m3u8"] = func(outDir, fileName, fileExt string) FileNameBuilder {
		return m3u8NameBuilder{

			outDir:   outDir,
			fileName: fileName,
			fileExt:  fileExt,
		}
	}
}

// createFileNameBuilder create a FileNameBuilder instance
func createFileNameBuilder(file string) FileNameBuilder {
	fileExt := filepath.Ext(file)

	dir, fileName := filepath.Split(file)

	fileName = strings.TrimRight(fileName, fileExt)

	factory, ok := builderFactory[strings.ToLower(fileExt)]

	var builder FileNameBuilder

	if ok {
		builder = factory(dir, fileName, fileExt)
	} else {
		builder = &mediaNameBuilder{
			outDir:   dir,
			fileName: fileName,
			fileExt:  fileExt,
		}
	}

	return builder

}

type mediaNameBuilder struct {
	// outDir output dir
	outDir string
	// fileName output file name
	fileName string
	// fileExt extention of output file
	fileExt string
}

// OutDir output dir
func (b mediaNameBuilder) OutDir() string {
	return b.outDir
}

// OutFile build file
func (b mediaNameBuilder) OutFile() string {
	return filepath.Join(b.outDir, b.fileName+"%d"+b.fileExt)
}

// FileExt get file ext
func (b mediaNameBuilder) FileExt() string {
	return b.fileExt
}

// ClipsFile build filename
func (b mediaNameBuilder) ClipsFile(index int) string {
	return filepath.Join(b.outDir, b.ClipsFileName(index))
}

// ClipsFileName build filename
func (b mediaNameBuilder) ClipsFileName(index int) string {
	return b.fileName + strconv.Itoa(index) + b.fileExt
}

type m3u8NameBuilder struct {
	// outDir output dir
	outDir string
	// fileName output file name
	fileName string
	// fileExt extention of output file
	fileExt string
}

// OutDir output dir
func (b m3u8NameBuilder) OutDir() string {
	return b.outDir
}

// File build file
func (b m3u8NameBuilder) OutFile() string {
	return filepath.Join(b.outDir, b.fileName+b.fileExt)
}

func (b m3u8NameBuilder) ClipsFile(index int) string {
	return filepath.Join(b.outDir, b.ClipsFileName(index))
}

func (b m3u8NameBuilder) ClipsFileName(index int) string {
	return b.fileName + strconv.Itoa(index) + ".ts"
}

// FileExt get file ext
func (b m3u8NameBuilder) FileExt() string {
	return b.fileExt
}
