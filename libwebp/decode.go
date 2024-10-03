package libwebp

import (
	"image"
	"image/color"
	"io"

	"github.com/ryex/gowebp/libwebp/webpoptions"

	"github.com/ryex/gowebp/internal/libwebp"
)

// Encode encodes src as Webp into w using the options in o.
//
// Any src that isn't one of *image.RGBA, *image.NRGBA, or *image.Gray
// will be converted to *image.NRGBA using draw.Draw first.
func Decode(r io.Reader, o webpoptions.DecodingOptions) (img image.Image, err error) {
	var dec *libwebp.Decoder
	if dec, err = libwebp.NewDecoder(r, &o); err != nil {
		return
	}
	img, err = dec.Decode()
	return
}

func DecodeConfig(r io.Reader, o webpoptions.DecodingOptions) (image.Config, error) {
	if dec, err := libwebp.NewDecoder(r, &o); err != nil {
		return image.Config{}, err
	} else {
		return image.Config{
			ColorModel: color.RGBAModel,
			Width:      dec.GetFeatures().Width,
			Height:     dec.GetFeatures().Height,
		}, nil
	}
}
