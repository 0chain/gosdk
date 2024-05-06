package imaging

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	_ "github.com/gen2brain/heic"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/ccitt"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vector"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
)

type ResampleFilter string

const (
	// NearestNeighbor is a nearest-neighbor filter (no anti-aliasing).
	NearestNeighbor ResampleFilter = "NearestNeighbor"
	// Box filter (averaging pixels).
	Box = "Box"
	// Linear filter.
	Linear = "Linear"
	// Hermite cubic spline filter (BC-spline; B=0; C=0).
	Hermite = "Hermite"
	// MitchellNetravali is Mitchell-Netravali cubic filter (BC-spline; B=1/3; C=1/3).
	MitchellNetravali = "MitchellNetravali"
	// CatmullRom is a Catmull-Rom - sharp cubic filter (BC-spline; B=0; C=0.5).
	CatmullRom = "CatmullRom"
	// BSpline is a smooth cubic filter (BC-spline; B=1; C=0).
	BSpline = "BSpline"
	// Gaussian is a Gaussian blurring filter.
	Gaussian = "Gaussian"
	// Bartlett is a Bartlett-windowed sinc filter (3 lobes).
	Bartlett = "Bartlett"
	// Lanczos filter (3 lobes).
	Lanczos = "Lanczos"
	// Hann is a Hann-windowed sinc filter (3 lobes).
	Hann = "Hann"
	// Hamming is a Hamming-windowed sinc filter (3 lobes).
	Hamming = "Hamming"
	// Blackman is a Blackman-windowed sinc filter (3 lobes).
	Blackman = "Blackman"
	// Welch is a Welch-windowed sinc filter (parabolic window, 3 lobes).
	Welch = "Welch"
	// Cosine is a Cosine-windowed sinc filter (3 lobes).
	Cosine = "Cosine"
)

var resample map[ResampleFilter]imaging.ResampleFilter

func init() {
	resample = map[ResampleFilter]imaging.ResampleFilter{
		NearestNeighbor:   imaging.NearestNeighbor,
		Box:               imaging.Box,
		Linear:            imaging.Linear,
		Hermite:           imaging.Hermite,
		MitchellNetravali: imaging.MitchellNetravali,
		CatmullRom:        imaging.CatmullRom,
		BSpline:           imaging.BSpline,
		Gaussian:          imaging.Gaussian,
		Bartlett:          imaging.Bartlett,
		Lanczos:           imaging.Lanczos,
		Hann:              imaging.Hann,
		Hamming:           imaging.Hamming,
		Blackman:          imaging.Blackman,
		Welch:             imaging.Welch,
		Cosine:            imaging.Cosine,
	}

}

func Thumbnail(buf []byte, width, height int, filter ResampleFilter) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	filterValue := imaging.Lanczos
	if fv, has := resample[filter]; has {
		filterValue = fv
	}
	nrgba := imaging.Thumbnail(img, width, height, filterValue)
	fd := &bytes.Buffer{}
	err = jpeg.Encode(fd, nrgba, nil)
	if err != nil {
		return nil, err
	}
	return fd.Bytes(), nil
}
