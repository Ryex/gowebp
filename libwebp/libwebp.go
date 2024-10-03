package libwebp

import (
	"image"
	"io"

	"github.com/ryex/gowebp/libwebp/webpoptions"
)

func init() {
	image.RegisterFormat("webp", "RIFF????WEBPVP8", quickDecode, quickDecodeConfig)
}

func quickDecode(r io.Reader) (image.Image, error) {
	return Decode(r, webpoptions.DecodingOptions{})
}

func quickDecodeConfig(r io.Reader) (image.Config, error) {
	return DecodeConfig(r, webpoptions.DecodingOptions{})
}
