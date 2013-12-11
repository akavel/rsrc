package main

import (
	"debug/pe"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
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

type RelocationEntry struct {
	RVA         uint32 // "offset within the Section's raw data where the address starts."
	SymbolIndex uint32 // "(zero based) index in the Symbol table to which the reference refers."
	Type        uint16
}

type Symbol struct {
	Name           [8]byte
	Value          uint32
	SectionNumber  uint16
	Type           uint16
	StorageClass   uint8
	AuxiliaryCount uint8
}

type StringsHeader struct {
	Length uint32
}

const (
	MASK_SUBDIRECTORY = 1 << 31
	TYPE_MANIFEST     = 24
)

var (
	STRING_RSRC = [8]byte{'.', 'r', 's', 'r', 'c', 0, 0, 0}
)

func MustGetFieldOffset(t reflect.Type, field string) uintptr {
	f, ok := t.FieldByName(field)
	if !ok {
		panic("field " + field + " not found")
	}
	return f.Offset
}

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
	w.Offset += uint32(n)
}

type SizedReader interface {
	io.Reader
	Size() int64
}

func main() {
	//TODO: allow in options advanced specification of multiple resources, as a tree (json?)
	var fnamein, fnameout string
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&fnamein, "manifest", "", "path to a Windows manifest file to embed")
	flags.StringVar(&fnameout, "o", "rsrc.syso", "name of output COFF (.res or .syso) file")
	_ = flags.Parse(os.Args[1:])
	if fnamein == "" {
		fmt.Fprintf(os.Stderr, "USAGE: %s -manifest FILE.exe.manifest [-o FILE.syso]\n"+
			"Generates a .syso file with specified resources embedded in .rsrc section,\n"+
			"aimed for consumption by Go linker when building Win32 excecutables.\n"+
			"OPTIONS:\n",
			os.Args[0])
		flags.PrintDefaults()
		os.Exit(1)
	}

	err := run(fnamein, fnameout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(fnamein, fnameout string) error {
	var manifest SizedReader
	{
		f, err := os.Open(fnamein)
		if err != nil {
			return fmt.Errorf("Error opening manifest file '%s': %s", fnamein, err)
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return err
		}
		manifest = io.NewSectionReader(f, 0, info.Size())
	}

	out, err := os.Create(fnameout)
	if err != nil {
		return err
	}
	defer out.Close()
	w := Writer{W: out}

	// precalculate some important offsets in resulting file, that we must know earlier
	//TODO: try to simplify by adding fake section at beginning, containing strings table in data, and characteristics saying "drop me when linking"
	rawdataoff := uint32(binary.Size(pe.FileHeader{}) + binary.Size(pe.SectionHeader32{}))
	hierarchylen := uint32(3*binary.Size(ImageResourceDirectory{}) +
		3*binary.Size(ImageResourceDirectoryEntry{}))
	rawdatalen := hierarchylen +
		uint32(1*binary.Size(ImageResourceDataEntry{})) +
		uint32(manifest.Size())
	diroff := rawdataoff
	relocoff := rawdataoff + rawdatalen
	relocp := hierarchylen + uint32(MustGetFieldOffset(reflect.TypeOf(ImageResourceDataEntry{}), "OffsetToData"))
	reloclen := uint32(binary.Size(RelocationEntry{}))
	symoff := relocoff + reloclen

	w.WriteLE(pe.FileHeader{
		Machine:              0x014c, //FIXME: find out how to differentiate this value, or maybe not necessary for Go
		NumberOfSections:     1,      // .rsrc
		TimeDateStamp:        0,      // was also 0 in sample data from MinGW's windres.exe
		PointerToSymbolTable: uint32(symoff),
		NumberOfSymbols:      1,
		SizeOfOptionalHeader: 0,
		Characteristics:      0x0104, //FIXME: copied from windres.exe output, find out what should be here and why
	})
	w.WriteLE(pe.SectionHeader32{
		Name:                 STRING_RSRC,
		SizeOfRawData:        rawdatalen,
		PointerToRawData:     rawdataoff,
		PointerToRelocations: relocoff,
		NumberOfRelocations:  1,
		Characteristics:      0x40000040, // "INITIALIZED_DATA MEM_READ" ?
	})

	// now, build "directory hierarchy" of .rsrc section: first type, then id/name, then language
	w.WriteLE(ImageResourceDirectory{
		NumberOfIdEntries: 1,
	})
	w.WriteLE(ImageResourceDirectoryEntry{
		NameOrId:     TYPE_MANIFEST,
		OffsetToData: MASK_SUBDIRECTORY | (w.Offset + uint32(binary.Size(ImageResourceDirectoryEntry{})) - diroff),
	})
	w.WriteLE(ImageResourceDirectory{
		NumberOfIdEntries: 1,
	})
	w.WriteLE(ImageResourceDirectoryEntry{
		NameOrId:     1, // ID
		OffsetToData: MASK_SUBDIRECTORY | (w.Offset + uint32(binary.Size(ImageResourceDirectoryEntry{})) - diroff),
	})
	w.WriteLE(ImageResourceDirectory{
		NumberOfIdEntries: 1,
	})
	w.WriteLE(ImageResourceDirectoryEntry{
		NameOrId:     0x0409, //FIXME: language; what value should be here?
		OffsetToData: w.Offset + uint32(binary.Size(ImageResourceDirectoryEntry{})) - diroff,
	})

	w.WriteLE(ImageResourceDataEntry{
		OffsetToData: w.Offset + uint32(binary.Size(ImageResourceDataEntry{})) - diroff,
		Size1:        uint32(manifest.Size()),
		CodePage:     0, //FIXME: what value here? for now just tried 0
	})

	w.WriteFromSized(manifest)

	w.WriteLE(RelocationEntry{
		RVA:         relocp, // FIXME: IIUC, this resolves to value contained in ImageResourceDataEntry.OffsetToData
		SymbolIndex: 0,      // "(zero based) index in the Symbol table to which the reference refers. Once you have loaded the COFF file into memory and know where each symbol is, you find the new updated address for the given symbol and update the reference accordingly."
		Type:        7,      // according to ldpe.c, this decodes to: IMAGE_REL_I386_DIR32NB
	})

	w.WriteLE(Symbol{
		Name:           STRING_RSRC,
		Value:          0,
		SectionNumber:  1,
		Type:           0, // FIXME: wtf?
		StorageClass:   3, // FIXME: is it ok? and uint8? and what does the value mean?
		AuxiliaryCount: 0, // FIXME: wtf?
	})

	w.WriteLE(StringsHeader{
		Length: uint32(binary.Size(StringsHeader{})), // empty strings table -- but we must still show size of the table's header...
	})

	if w.Err != nil {
		return fmt.Errorf("Error writing .rsrc Symbol Table & Strings: %s", w.Err)
	}

	return nil
}
