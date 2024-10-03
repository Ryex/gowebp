package libwebp

/*
#include <stdlib.h>
#include <string.h> // for memset
#ifndef LIBWEBP_NO_SRC
#include "decode.h"
#else
#include <webp/encode.h>
#endif
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"io"
	"unsafe"

	"github.com/ryex/gowebp/libwebp/webpoptions"
)

// noinspection GoUnusedConst
const (
	Vp8StatusOk VP8StatusCode = iota
	Vp8StatusOutOfMemory
	Vp8StatusInvalidParam
	Vp8StatusBitstreamError
	Vp8StatusUnsupportedFeature
	Vp8StatusSuspended
	Vp8StatusUserAbort
	Vp8StatusNotEnoughData
)

type VP8StatusCode int

func (c VP8StatusCode) String() (label string) {
	switch c {
	case Vp8StatusOk:
		label = "VP8_STATUS_OK"
	case Vp8StatusOutOfMemory:
		label = "VP8_STATUS_OUT_OF_MEMORY"
	case Vp8StatusInvalidParam:
		label = "VP8_STATUS_INVALID_PARAM"
	case Vp8StatusBitstreamError:
		label = "VP8_STATUS_BITSTREAM_ERROR"
	case Vp8StatusUnsupportedFeature:
		label = "VP8_STATUS_UNSUPPORTED_FEATURE"
	case Vp8StatusSuspended:
		label = "VP8_STATUS_SUSPENDED"
	case Vp8StatusUserAbort:
		label = "VP8_STATUS_USER_ABORT"
	case Vp8StatusNotEnoughData:
		label = "VP8_STATUS_NOT_ENOUGH_DATA"
	default:
		label = "VP8 undefined status code"
	}

	return
}

const (
	Vp8EncOk Vp8EncStatus = iota
	Vp8EncErrorOutOfMemory
	Vp8EncErrorBitstreamOutOfMemory
	Vp8EncErrorNullParameter
	Vp8EncErrorInvalidConfiguration
	Vp8EncErrorBadDimension
	Vp8EncErrorPartition0Overflow
	Vp8EncErrorPartitionOverflow
	Vp8EncErrorBadWrite
	Vp8EncErrorFileTooBig
	Vp8EncErrorUserAbort
	Vp8EncErrorLast
)

type Vp8EncStatus int

func (c Vp8EncStatus) String() (label string) {
	switch c {
	case Vp8EncOk:
		label = "VP8_ENC_OK"
	case Vp8EncErrorOutOfMemory:
		label = "VP8_ENC_ERROR_OUT_OF_MEMORY"
	case Vp8EncErrorBitstreamOutOfMemory:
		label = "VP8_ENC_ERROR_BITSTREAM_OUT_OF_MEMORY"
	case Vp8EncErrorNullParameter:
		label = "VP8_ENC_ERROR_NULL_PARAMETER"
	case Vp8EncErrorInvalidConfiguration:
		label = "VP8_ENC_ERROR_INVALID_CONFIGURATION"
	case Vp8EncErrorBadDimension:
		label = "VP8_ENC_ERROR_BAD_DIMENSION"
	case Vp8EncErrorPartition0Overflow:
		label = "VP8_ENC_ERROR_PARTITION0_OVERFLOW"
	case Vp8EncErrorPartitionOverflow:
		label = "VP8_ENC_ERROR_PARTITION_OVERFLOW"
	case Vp8EncErrorBadWrite:
		label = "VP8_ENC_ERROR_BAD_WRITE"
	case Vp8EncErrorFileTooBig:
		label = "VP8_ENC_ERROR_FILE_TOO_BIG"
	case Vp8EncErrorUserAbort:
		label = "VP8_ENC_ERROR_USER_ABORT"
	case Vp8EncErrorLast:
		label = "VP8_ENC_ERROR_LAST"
	default:
		label = "VP8 undefined status code"
	}

	return
}

type FormatType int

// noinspection GoUnusedConst
const (
	FormatUndefined FormatType = iota
	FormatLossy
	FormatLossless
)

type BitstreamFeatures struct {
	Width        int
	Height       int
	HasAlpha     bool
	HasAnimation bool
	Format       FormatType
}

// Decoder stores information to decode picture
type Decoder struct {
	data    []byte
	options *webpoptions.DecodingOptions
	config  *C.WebPDecoderConfig
	dPtr    *C.uint8_t
	sPtr    C.size_t
}

func NewDecoder(r io.Reader, o *webpoptions.DecodingOptions) (d *Decoder, err error) {
	var data []byte

	if data, err = io.ReadAll(r); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	if o == nil {
		o = &webpoptions.DecodingOptions{}
	}
	d = &Decoder{data: data, options: o}

	if d.config, err = decodingOPtionsToCConfig(o); err != nil {
		return nil, err
	}

	d.dPtr = (*C.uint8_t)(&d.data[0])
	d.sPtr = (C.size_t)(len(d.data))

	if status := d.parseFeatures(d.dPtr, d.sPtr); status != Vp8StatusOk {
		return nil, fmt.Errorf("cannot fetch features: %s", status.String())
	}

	return
}

// Decode picture from reader
func (d *Decoder) Decode() (image.Image, error) {
	// вписываем размеры итоговой картинки
	d.config.output.width, d.config.output.height = d.getOutputDimensions()
	// указываем что декодируем в RGBA
	d.config.output.colorspace = C.MODE_RGBA
	d.config.output.is_external_memory = 1

	img := image.NewNRGBA(image.Rectangle{Max: image.Point{
		X: int(d.config.output.width),
		Y: int(d.config.output.height),
	}})

	buff := (*C.WebPRGBABuffer)(unsafe.Pointer(&d.config.output.u[0]))
	buff.stride = C.int(img.Stride)
	buff.rgba = (*C.uint8_t)(&img.Pix[0])
	buff.size = (C.size_t)(len(img.Pix))

	if status := VP8StatusCode(C.WebPDecode(d.dPtr, d.sPtr, d.config)); status != Vp8StatusOk {
		return nil, fmt.Errorf("cannot decode picture: %s", status.String())
	}

	return img, nil
}

// GetFeatures return information about picture: width, height ...
func (d *Decoder) GetFeatures() BitstreamFeatures {
	return BitstreamFeatures{
		Width:        int(d.config.input.width),
		Height:       int(d.config.input.height),
		HasAlpha:     int(d.config.input.has_alpha) == 1,
		HasAnimation: int(d.config.input.has_animation) == 1,
		Format:       FormatType(d.config.input.format),
	}
}

// parse features from picture
func (d *Decoder) parseFeatures(dataPtr *C.uint8_t, sizePtr C.size_t) VP8StatusCode {
	return VP8StatusCode(C.WebPGetFeatures(dataPtr, sizePtr, &d.config.input))
}

// return dimensions of result image
func (d *Decoder) getOutputDimensions() (width, height C.int) {
	width = d.config.input.width
	height = d.config.input.height

	if d.config.options.use_scaling > 0 {
		width = d.config.options.scaled_width
		height = d.config.options.scaled_height
	} else if d.config.options.use_cropping > 0 {
		width = d.config.options.crop_width
		height = d.config.options.crop_height
	}

	return
}

func decodingOPtionsToCConfig(o *webpoptions.DecodingOptions) (*C.WebPDecoderConfig, error) {
	cfg := &C.WebPDecoderConfig{}
	if C.WebPInitDecoderConfig(cfg) == 0 {
		return nil, errors.New("failed to init decode config")
	}

	if o.BypassFiltering {
		cfg.options.bypass_filtering = 1
	}

	if o.NoFancyUpsampling {
		cfg.options.no_fancy_upsampling = 1
	}

	// проверяем надо ли кропнуть
	if o.Crop.Max.X > 0 && o.Crop.Max.Y > 0 {
		cfg.options.use_cropping = 1
		cfg.options.crop_left = C.int(o.Crop.Min.X)
		cfg.options.crop_top = C.int(o.Crop.Min.Y)
		cfg.options.crop_width = C.int(o.Crop.Max.X)
		cfg.options.crop_height = C.int(o.Crop.Max.Y)
	}

	// проверяем надо ли заскейлить
	if o.Scale.Max.X > 0 && o.Scale.Max.Y > 0 {
		cfg.options.use_scaling = 1
		cfg.options.scaled_width = C.int(o.Scale.Max.X)
		cfg.options.scaled_height = C.int(o.Scale.Max.Y)
	}

	if o.UseThreads {
		cfg.options.use_threads = 1
	}

	cfg.options.dithering_strength = C.int(o.DitheringStrength)

	if o.Flip {
		cfg.options.flip = 1
	}

	cfg.options.alpha_dithering_strength = C.int(o.AlphaDitheringStrength)

	return cfg, nil
}
