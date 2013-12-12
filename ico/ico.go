// Package ico describes Windows ICO file format.
package ico

// ICO: http://msdn.microsoft.com/en-us/library/ms997538.aspx
// BMP/DIB: http://msdn.microsoft.com/en-us/library/windows/desktop/dd183562%28v=vs.85%29.aspx

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
)

const (
	BI_RGB = 0
)

type ICONDIR struct {
	Reserved uint16 // must be 0
	Type     uint16 // Resource Type (1 for icons)
	Count    uint16 // How many images?
}

type ICONDIRENTRY struct {
	Width       byte   // Width, in pixels, of the image
	Height      byte   // Height, in pixels, of the image
	ColorCount  byte   // Number of colors in image (0 if >=8bpp)
	Reserved    byte   // Reserved (must be 0)
	Planes      uint16 // Color Planes
	BitCount    uint16 // Bits per pixel
	BytesInRes  uint32 // How many bytes in this resource?
	ImageOffset uint32 // Where in the file is this image? [from beginning of file]
}

type BITMAPINFOHEADER struct {
	Size          uint32
	Width         int32
	Height        int32  // NOTE: "represents the combined height of the XOR and AND masks. Remember to divide this number by two before using it to perform calculations for either of the XOR or AND masks."
	Planes        uint16 // [BMP/DIB]: "is always 1"
	BitCount      uint16
	Compression   uint32 // for ico = 0
	SizeImage     uint32
	XPelsPerMeter int32  // for ico = 0
	YPelsPerMeter int32  // for ico = 0
	ClrUsed       uint32 // for ico = 0
	ClrImportant  uint32 // for ico = 0
}

type RGBQUAD struct {
	Blue     byte
	Green    byte
	Red      byte
	Reserved byte // must be 0
}

func skip(r io.Reader, n int64) error {
	_, err := io.CopyN(ioutil.Discard, r, n)
	return err
}

type icoOffset struct {
	n      int
	offset uint32
}

type rawico struct {
	icoinfo ICONDIRENTRY
	bmpinfo *BITMAPINFOHEADER
	idx     int
	data    []byte
}

type byOffsets []rawico

func (o byOffsets) Len() int           { return len(o) }
func (o byOffsets) Less(i, j int) bool { return o[i].icoinfo.ImageOffset < o[j].icoinfo.ImageOffset }
func (o byOffsets) Swap(i, j int) {
	tmp := o[i]
	o[i] = o[j]
	o[j] = tmp
}

type ICO struct{}

// NOTE: won't succeed on files with overlapping offsets
func DecodeAll(r io.Reader) ([]*ICO, error) {
	var hdr ICONDIR
	err := binary.Read(r, binary.LittleEndian, &hdr)
	if err != nil {
		return nil, err
	}
	if hdr.Reserved != 0 || hdr.Type != 1 {
		return nil, fmt.Errorf("bad magic number")
	}

	raws := make([]rawico, hdr.Count)
	for i := 0; i < len(raws); i++ {
		err = binary.Read(r, binary.LittleEndian, &raws[i].icoinfo)
		if err != nil {
			return nil, err
		}
		raws[i].idx = i
	}

	sort.Sort(byOffsets(raws))

	offset := uint32(binary.Size(&hdr) + len(raws)*binary.Size(ICONDIRENTRY{}))
	for i := 0; i < len(raws); i++ {
		err = skip(r, int64(raws[i].icoinfo.ImageOffset-offset))
		if err != nil {
			return nil, err
		}
		offset = raws[i].icoinfo.ImageOffset

		raws[i].bmpinfo = &BITMAPINFOHEADER{}
		err = binary.Read(r, binary.LittleEndian, raws[i].bmpinfo)
		if err != nil {
			return nil, err
		}

		raws[i].data = make([]byte, raws[i].icoinfo.BytesInRes-uint32(binary.Size(BITMAPINFOHEADER{})))
		_, err = io.ReadFull(r, raws[i].data)
		if err != nil {
			return nil, err
		}
	}

	icos := make([]*ICO, len(raws))
	for i := 0; i < len(raws); i++ {
		icos[raws[i].idx], err = decode(raws[i].bmpinfo, raws[i].data)
		if err != nil {
			return nil, err
		}
	}
	return icos, nil
}

func decode(info *BITMAPINFOHEADER, data []byte) (*ICO, error) {
	bottomup := info.Height > 0
	if !bottomup {
		info.Height = -info.Height
	}

	if info.Compression != BI_RGB {
		return nil, fmt.Errorf("ICO compression not supported (got %d)", info.Compression)
	}

	switch info.BitCount {
	default:
		return nil, fmt.Errorf("unsupported ICO bit depth (BitCount) %d", info.BitCount)
	}
}
