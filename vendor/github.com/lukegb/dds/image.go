/*
Copyright 2017 Luke Granger-Brown

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package dds provides a decoder for the DirectDraw surface format,
// which is compatible with the standard image package.
//
// It should normally be used by importing it with a blank name, which
// will cause it to register itself with the image package:
//  import _ "github.com/lukegb/dds"
package dds

import (
	"fmt"
	"image"
	"image/color"
	"io"
)

func init() {
	image.RegisterFormat("dds", "DDS ", Decode, DecodeConfig)
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	h, err := readHeader(r)
	if err != nil {
		return image.Config{}, err
	}

	// set width and height
	c := image.Config{
		Width:  int(h.width),
		Height: int(h.height),
	}

	pf := h.pixelFormat
	hasAlpha := (pf.flags&pfAlphaPixels == pfAlphaPixels) || (pf.flags&pfAlpha == pfAlpha)
	hasRGB := (pf.flags&pfFourCC == pfFourCC) || (pf.flags&pfRGB == pfRGB)
	hasYUV := (pf.flags&pfYUV == pfYUV)
	hasLuminance := (pf.flags&pfLuminance == pfLuminance)
	switch {
	case hasRGB && pf.rgbBitCount == 32:
		c.ColorModel = color.RGBAModel
	case hasRGB && pf.rgbBitCount == 64:
		c.ColorModel = color.RGBA64Model
	case hasYUV && pf.rgbBitCount == 24:
		c.ColorModel = color.YCbCrModel
	case hasLuminance && pf.rgbBitCount == 8:
		c.ColorModel = color.GrayModel
	case hasLuminance && pf.rgbBitCount == 16:
		c.ColorModel = color.Gray16Model
	case hasAlpha && pf.rgbBitCount == 8:
		c.ColorModel = color.AlphaModel
	case hasAlpha && pf.rgbBitCount == 16:
		c.ColorModel = color.AlphaModel
	default:
		return image.Config{}, fmt.Errorf("unrecognized image format: hasAlpha: %v, hasRGB: %v, hasYUV: %v, hasLuminance: %v, pf.flags: %x", hasAlpha, hasRGB, hasYUV, hasLuminance, pf.flags)
	}

	return c, nil
}

func Decode(r io.Reader) (image.Image, error) {
	h, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	if h.pixelFormat.flags&pfFourCC == pfFourCC {
		return nil, fmt.Errorf("image data is compressed with %v; compression is unsupported", h.pixelFormat.fourCC)
	}

	if h.pixelFormat.flags&(pfAlphaPixels|pfRGB) != h.pixelFormat.flags {
		return nil, fmt.Errorf("unsupported pixel format %x", h.pixelFormat.flags)
	}
	iw, ih := int(h.width), int(h.height)
	im := image.NewRGBA(image.Rect(0, 0, iw, ih))
	st := int(h.pixelFormat.rgbBitCount / 8)
	if st*8 != int(h.pixelFormat.rgbBitCount) {
		return nil, fmt.Errorf("unsupported bit count: %d", h.pixelFormat.rgbBitCount)
	}
	rb := lowestSetBit(h.pixelFormat.rBitMask)
	gb := lowestSetBit(h.pixelFormat.gBitMask)
	bb := lowestSetBit(h.pixelFormat.bBitMask)
	ab := lowestSetBit(h.pixelFormat.aBitMask)
	if rb&7 != 0 || gb&7 != 0 || bb&7 != 0 || ab&7 != 0 {
		return nil, fmt.Errorf("unsupported bitmasks: %x %x %x %x", h.pixelFormat.rBitMask, h.pixelFormat.gBitMask, h.pixelFormat.bBitMask, h.pixelFormat.aBitMask)
	}
	rb /= 8
	gb /= 8
	bb /= 8
	ab /= 8
	buf := make([]byte, st*iw)
	px := im.Pix

	noalpha := h.pixelFormat.flags&pfAlphaPixels == 0

	for y := 0; y < ih; y++ {
		if _, err := r.Read(buf); err != nil {
			return nil, err
		}
		for x := 0; x < iw; x++ {
			px[y*im.Stride+x*4+0] = buf[x*st+rb]
			px[y*im.Stride+x*4+1] = buf[x*st+gb]
			px[y*im.Stride+x*4+2] = buf[x*st+bb]
			if noalpha {
				px[y*im.Stride+x*4+3] = 255
			} else {
				px[y*im.Stride+x*4+3] = buf[x*st+ab]
			}
		}
	}
	return im, nil
}
