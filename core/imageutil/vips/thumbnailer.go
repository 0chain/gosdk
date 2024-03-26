package vips

import (
	"github.com/davidbyttow/govips/v2/vips"
)

var crop map[Crop]vips.Interesting

func init() {
	crop = map[Crop]vips.Interesting{
		None:      vips.InterestingNone,
		Centre:    vips.InterestingCentre,
		Entropy:   vips.InterestingEntropy,
		Attention: vips.InterestingAttention,
		Low:       vips.InterestingLow,
		High:      vips.InterestingHigh,
		All:       vips.InterestingAll,
		Last:      vips.InterestingLast,
	}
}

type Crop string

const (
	None Crop = "None"
	Centre Crop = "Centre"
	Entropy Crop = "Entropy"
	Attention Crop = "Attention"
	Low Crop = "Low"
	High Crop = "High"
	All Crop = "All"
	Last Crop = "Last"
)

func Thumbnail(buf []byte, width, height int, crp Crop) ([]byte, error) {
	vipsImgRef, err := vips.NewImageFromBuffer(buf)
	if err != nil {
		return nil, err
	}
	cropValue := vips.InterestingAll
	if vipsI, has := crop[crp]; has {
		cropValue = vipsI
	}
	err = vipsImgRef.Thumbnail(width, height, cropValue)
	if err != nil {
		return nil, err
	}
	jpegBytes, _, err := vipsImgRef.ExportJpeg(vips.NewJpegExportParams())
	return jpegBytes, err
}