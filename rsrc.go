package main

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"unsafe"
)

type ImageResourceDirectory struct {
	Characteristics      uint32
	TimeDateStamp        uint32
	MajorVersion         uint16
	MinorVersion         uint16
	NumberOfNamedEntries uint16
	NumberOfIdEntries    uint16
}

type ImageResourceDirectoryEntry struct {
	NameOrId     uint32
	OffsetToData uint32
}

type ImageResourceDataEntry struct {
	OffsetToData uint32
	Size1        uint32
	CodePage     uint32
	Reserved     uint32
}

const (
	MASK_SUBDIRECTORY = 1 << 31
	TYPE_MANIFEST     = 24
)

type Writer struct {
	W      io.Writer
	Offset uint32 //FIXME: uint64?
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

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	//TODO: allow options:
	// -o FILENAME - output file name
	// + advanced specification of multiple resources, as a tree (json?)
	if len(os.Args) <= 1 {
		return fmt.Errorf("USAGE: %s FILE.exe.manifest\n"+
			"Generates FILE.res",
			os.Args[0])
	}

	fname := os.Args[1]
	suffix := ".exe.manifest"
	if !strings.HasSuffix(fname, suffix) {
		return fmt.Errorf("Filename '%s' does not end in suffix '%s'", fname, suffix)
	}
	fname = fname[:len(fname)-len(suffix)]

	manifest, err := ioutil.ReadFile(fname + suffix)
	if err != nil {
		return err
	}

	out, err := os.Create(fname + ".res")
	if err != nil {
		return err
	}
	defer out.Close()
	w := Writer{W: out}

	coffhdr := pe.FileHeader{
		Machine:              0x014c, //FIXME: find out how to differentiate this value, or maybe not necessary for Go
		NumberOfSections:     1,      // .rsrc
		TimeDateStamp:        0,      // was also 0 in sample data from MinGW's windres.exe
		PointerToSymbolTable: 0,
		NumberOfSymbols:      0,
		SizeOfOptionalHeader: 0,
		Characteristics:      0x0104, //FIXME: copied from windres.exe output, find out what should be here and why
	}
	w.WriteLE(coffhdr)
	if w.Err != nil {
		return fmt.Errorf("Error writing COFF header: %s", w.Err)
	}

	secthdr := pe.SectionHeader32{
		Name:             [8]byte{'.', 'r', 's', 'r', 'c', 0, 0, 0},
		SizeOfRawData:    uint32(len(manifest)), //FIXME: probably must include all the .rsrc directory structures too
		PointerToRawData: w.Offset + uint32(unsafe.Sizeof(pe.SectionHeader32{})),
		Characteristics:  0x40000040, // "INITIALIZED_DATA MEM_READ" ?
	}
	w.WriteLE(secthdr)
	if w.Err != nil {
		return fmt.Errorf("Error writing .rsrc section header: %s", w.Err)
	}

	// now, build "directory hierarchy" of .rsrc section: first type, then id/name, then language

	diroff := w.Offset // all "OffsetToData" are relative to this point

	w.WriteLE(ImageResourceDirectory{
		NumberOfIdEntries: 1,
	})
	w.WriteLE(ImageResourceDirectoryEntry{
		NameOrId:     TYPE_MANIFEST,
		OffsetToData: MASK_SUBDIRECTORY | (w.Offset + uint32(unsafe.Sizeof(ImageResourceDirectoryEntry{})) - diroff),
	})
	w.WriteLE(ImageResourceDirectory{
		NumberOfIdEntries: 1,
	})
	w.WriteLE(ImageResourceDirectoryEntry{
		NameOrId:     1, // ID
		OffsetToData: MASK_SUBDIRECTORY | (w.Offset + uint32(unsafe.Sizeof(ImageResourceDirectoryEntry{})) - diroff),
	})
	w.WriteLE(ImageResourceDirectory{
		NumberOfIdEntries: 1,
	})
	w.WriteLE(ImageResourceDirectoryEntry{
		NameOrId:     0x0409, //FIXME: language; what value should be here?
		OffsetToData: w.Offset + uint32(unsafe.Sizeof(ImageResourceDirectoryEntry{})) - diroff,
	})

	w.WriteLE(ImageResourceDataEntry{
		OffsetToData: w.Offset + uint32(unsafe.Sizeof(ImageResourceDataEntry{})) - diroff,
		Size1:        uint32(len(manifest)),
		CodePage:     0, //FIXME: what value here? for now just tried 0
	})

	if w.Err != nil {
		return fmt.Errorf("Error writing .rsrc Directory Hierarchy: %s", w.Err)
	}

	_, err = w.W.Write(manifest)
	if err != nil {
		return fmt.Errorf("Error writing manifest contents: %s", err)
	}

	return nil
}
