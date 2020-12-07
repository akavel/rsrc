package binutil

import (
	"encoding/binary"
	"io"
	"reflect"
)

var pad [8]byte

type Writer struct {
	W      io.Writer
	Offset uint32 //FIXME: int64?
	Err    error
}

func (w *Writer) WriteLE(v interface{}) {
	if w.Err != nil {
		return
	}
	w.Err = binary.Write(w.W, binary.LittleEndian, v)
	if w.Err != nil {
		return
	}
	w.Offset += uint32(reflect.TypeOf(v).Size())
}

func (w *Writer) WriteFromSized(r SizedReader) {
	if w.Err != nil {
		return
	}
	var n int64
	n, w.Err = io.CopyN(w.W, r, r.Size())
	if w.Err != nil {
		return
	}
	aligned := RoomTaken(r)
	if aligned > n {
		w.W.Write(pad[:aligned-n])
		n = aligned
	}
	w.Offset += uint32(n)
}
