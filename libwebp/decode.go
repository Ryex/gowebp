package libwebp

import (
	"image"
	"io"

	"github.com/bep/gowebp/libwebp/webpoptions"

	"github.com/bep/gowebp/internal/libwebp"
)

// Encode encodes src as Webp into w using the options in o.
//
// Any src that isn't one of *image.RGBA, *image.NRGBA, or *image.Gray
// will be converted to *image.NRGBA using draw.Draw first.
func Decode(r io.Reader, o webpoptions.DecodingOptions) (img image.Image, err error) {
	var d *libwebp.Decoder
	if d, err = libwebp.NewDecoder(r, &o); err != nil {
		return
	}
	img, err = d.Decode()
	return
}
