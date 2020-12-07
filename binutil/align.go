package binutil

type Sizer interface {
	Size() int64
}

type AlignedSizer interface {
	Size() int64
	AlignedSize() int64
}

func Align(s int64) int64 {
	return (s-1)&^7 + 8
}

func RoomTaken(s Sizer) int64 {
	if as, ok := s.(AlignedSizer); ok {
		return as.AlignedSize()
	}
	return s.Size()
}
