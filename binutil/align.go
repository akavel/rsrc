package binutil

import "io"

type Sizer interface {
	Size() int64
}

type AlignedSectionReader struct {
	*io.SectionReader
}

func (a AlignedSectionReader) AlignedSize() int64 {
	return Align(a.Size())
}

type AlignedSizer interface {
	Size() int64
	AlignedSize() int64
}

func Align(s int64) int64 {
	return (s + 7) &^ 7
}

func RoomTaken(s Sizer) int64 {
	if as, ok := s.(AlignedSizer); ok {
		return as.AlignedSize()
	}
	return s.Size()
}
