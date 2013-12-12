// Package ico describes Windows ICO file format.
package ico

// http://msdn.microsoft.com/en-us/library/ms997538.aspx

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
	Height        int32 // NOTE: "represents the combined height of the XOR and AND masks. Remember to divide this number by two before using it to perform calculations for either of the XOR or AND masks."
	Planes        uint16
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

type icoOffsets []icoOffset

func (o icoOffsets) Len() int           { return len(o) }
func (o icoOffsets) Less(i, j int) bool { return o[i].offset < o[j].offset }
func (o icoOffsets) Swap(i, j int) {
	tmp := o[i]
	o[i] = o[j]
	o[j] = tmp
}

type rawico struct {
	icoinfo ICONDIRENTRY
	bmpinfo BITMAPINFOHEADER
	idx     int
	data    []byte
}

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

	entries := make([]ICONDIRENTRY, hdr.Count)
	offsets := make([]icoOffset, hdr.Count)
	for i := 0; i < len(entries); i++ {
		err = binary.Read(r, binary.LittleEndian, &entries[i])
		if err != nil {
			return nil, err
		}
		offsets[i] = icoOffset{i, entries[i].ImageOffset}
	}

	sort.Sort(icoOffsets(offsets))

	datas := make([][]byte, hdr.Count)
	offset := binary.Size(&hdr) + len(entries)*binary.Size(ICONDIRENTRY{})
	for i := 0; i < len(offsets); i++ {
		err = skip(r, offsets[i].offset-offset)
		if err != nil {
			return nil, err
		}
		offset = offsets[i].offset

		datas[i] = make([]byte, entries[offsets[i].n].BytesInRes)
		_, err = io.ReadFull(r, datas[i])
		if err != nil {
			return nil, err
		}
	}
}
