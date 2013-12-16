package main

// TODO:
// - to store icons, see: http://blogs.msdn.com/b/oldnewthing/archive/2012/07/20/10331787.aspx
//   - also need to first read and split input ico file

import (
	"debug/pe"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"

	"github.com/akavel/rsrc/ico"
)

type Dir struct { // struct IMAGE_RESOURCE_DIRECTORY
	Characteristics      uint32
	TimeDateStamp        uint32
	MajorVersion         uint16
	MinorVersion         uint16
	NumberOfNamedEntries uint16
	NumberOfIdEntries    uint16
	DirEntries
	Dirs
}

type DirEntries []DirEntry
type Dirs []Dir

type DirEntry struct { // struct IMAGE_RESOURCE_DIRECTORY_ENTRY
	NameOrId     uint32
	OffsetToData uint32
}

type DataEntry struct { // struct IMAGE_RESOURCE_DATA_ENTRY
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

	RT_ICON       = 3
	RT_GROUP_ICON = 3 + 11
	RT_MANIFEST   = 24
)

var (
	STRING_RSRC = [8]byte{'.', 'r', 's', 'r', 'c', 0, 0, 0}
	LANG_ENTRY  = DirEntries{{NameOrId: 0x0409}} //FIXME: language; what value should be here?
)

type GRPICONDIR struct {
	ico.ICONDIR
	Entries []GRPICONDIRENTRY
}

type GRPICONDIRENTRY struct {
	ico.IconDirEntryCommon
	Id uint16
}

func main() {
	//TODO: allow in options advanced specification of multiple resources, as a tree (json?)
	var fnamein, fnameico, fnameout string
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&fnamein, "manifest", "", "path to a Windows manifest file to embed")
	flags.StringVar(&fnameico, "ico", "", "UNSUPPORTED: path to ICO file to embed")
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

	err := run(fnamein, fnameico, fnameout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Coff struct {
	pe.FileHeader
	pe.SectionHeader32

	Dir
	DataEntries []DataEntry
	Data        []interface{}

	Relocations []RelocationEntry
	Symbols     []Symbol
	StringsHeader
}

func run(fnamein, fnameico, fnameout string) error {
	manifest, err := SizedOpen(fnamein)
	if err != nil {
		return fmt.Errorf("Error opening manifest file '%s': %s", fnamein, err)
	}
	defer manifest.Close()

	var icons []ico.ICONDIRENTRY
	var iconsf *os.File
	if fnameico != "" {
		iconsf, err = os.Open(fnameico)
		if err != nil {
			return err
		}
		defer iconsf.Close()
		icons, err = ico.DecodeHeaders(iconsf)
		if err != nil {
			return err
		}
	}

	newid := make(chan uint16)
	go func() {
		for i := uint16(1); ; i++ {
			newid <- i
		}
	}()

	out, err := os.Create(fnameout)
	if err != nil {
		return err
	}
	defer out.Close()
	w := Writer{W: out}

	coff := Coff{
		pe.FileHeader{
			Machine:              0x014c, //FIXME: find out how to differentiate this value, or maybe not necessary for Go
			NumberOfSections:     1,      // .rsrc
			TimeDateStamp:        0,      // was also 0 in sample data from MinGW's windres.exe
			NumberOfSymbols:      1,
			SizeOfOptionalHeader: 0,
			Characteristics:      0x0104, //FIXME: copied from windres.exe output, find out what should be here and why
		},
		pe.SectionHeader32{
			Name:                STRING_RSRC,
			NumberOfRelocations: 1,
			Characteristics:     0x40000040, // "INITIALIZED_DATA MEM_READ" ?
		},

		// now, build "directory hierarchy" of .rsrc section: first type, then id/name, then language
		Dir{
			NumberOfIdEntries: 1,
			DirEntries:        DirEntries{{NameOrId: RT_MANIFEST}},
			Dirs: Dirs{{
				NumberOfIdEntries: 1,
				DirEntries:        DirEntries{{NameOrId: uint32(<-newid)}}, // resource ID
				Dirs: Dirs{{
					NumberOfIdEntries: 1,
					DirEntries:        LANG_ENTRY,
				}},
			}},
		},
		[]DataEntry{
			DataEntry{
				Size1:    uint32(manifest.Size()),
				CodePage: 0, //FIXME: what value here? for now just tried 0
			},
		},
		[]interface{}{
			manifest,
		},

		[]RelocationEntry{RelocationEntry{
			SymbolIndex: 0, // "(zero based) index in the Symbol table to which the reference refers. Once you have loaded the COFF file into memory and know where each symbol is, you find the new updated address for the given symbol and update the reference accordingly."
			Type:        7, // according to ldpe.c, this decodes to: IMAGE_REL_I386_DIR32NB
		}},

		[]Symbol{Symbol{
			Name:           STRING_RSRC,
			Value:          0,
			SectionNumber:  1,
			Type:           0, // FIXME: wtf?
			StorageClass:   3, // FIXME: is it ok? and uint8? and what does the value mean?
			AuxiliaryCount: 0, // FIXME: wtf?
		}},

		StringsHeader{
			Length: uint32(binary.Size(StringsHeader{})), // empty strings table -- but we must still show size of the table's header...
		},
	}

	if len(icons) > 0 {
		//FIXME: add corresponding DataEntries
		//FIXME: add relocations
		/*
			coff.Dir.NumberOfIdEntries+=2

			coff.Dir.DirEntries = append(coff.Dir.DirEntries, DirEntry{NameOrId: RT_ICON})
			coff.Dir.Dirs = append(coff.Dir.Dirs, Dir{
				NumberOfIdEntries: len(icons),
			})
			dir := &coff.Dir.Dirs[len(coff.Dir.Dirs)-1]
			group := GRPICONDIR{
				Reserved: 0, // magic num.
				Type: 1, // magic num.
				Count: len(icons),
			}
			for i, icon := range icons {
				id := <-newid

				group.Entries = append(group.Entries, GRPICONDIRENTRY{icon, id})
				dir.DirEntries = append(dir.DirEntries, DirEntry{NameOrId: uint32(id)})
				dir.Dirs = append(dir.Dirs, Dir{
					NumberOfIdEntries: 1,
					DirEntries: LANG_ENTRY,
				})

				coff.Data = append(coff.Data, io.NewSectionReader(iconsf, icon.ImageOffset, icon.BytesInRes))
			}

			coff.Dir.DirEntries = append(coff.Dir.DirEntries, DirEntry{NameOrId: RT_GROUP_ICON})
			coff.Dir.Dirs = append(coff.Dir.Dirs, Dir{
				NumberOfIdEntries: 1,
				DirEntries: DirEntries{{NameOrId: uint32(<-newid)}},
				Dirs: Dirs{{
					NumberOfIdEntries: 1,
					DirEntries: LANG_ENTRY,
				}},
			})
			coff.Data = append(coff.Data, group)
		*/
	}

	leafwalker := make(chan *DirEntry)
	go func() {
		for _, dir1 := range coff.Dir.Dirs { // resource type
			for _, dir2 := range dir1.Dirs { // resource ID
				for i := range dir2.DirEntries { // resource lang
					leafwalker <- &dir2.DirEntries[i]
				}
			}
		}
	}()

	N := `\[(\d+)\]`
	dir_n := regexp.MustCompile("^/Dir/Dirs" + N + "$")
	dir_n_n := regexp.MustCompile("^/Dir/Dirs" + N + "/Dirs" + N + "$")
	dataentry_n := regexp.MustCompile("^/DataEntries" + N + "$")
	dataentry_n_off := regexp.MustCompile("^/DataEntries" + N + "/OffsetToData$")
	data_n := regexp.MustCompile("^/Data" + N + "$")

	// fill in some important offsets in resulting file
	var offset, diroff uint32
	Walk(coff, func(v reflect.Value, path string) error {
		switch path {
		case "/Dir":
			coff.SectionHeader32.PointerToRawData = offset
			diroff = offset
		//case "/Dir/Dirs[0]":
		//	coff.Dir.DirEntries[0].OffsetToData = MASK_SUBDIRECTORY | (offset - diroff)
		//case "/Dir/Dirs[0]/Dirs[0]":
		//	coff.Dir.Dirs[0].DirEntries[0].OffsetToData = MASK_SUBDIRECTORY | (offset - diroff)
		//case "/DataEntries[0]":
		//	direntry := <-leafwalker
		//	direntry.OffsetToData = offset - diroff
		//case "/DataEntries[0]/OffsetToData":
		//	coff.Relocations[0].RVA = offset - diroff
		//case "/Data[0]":
		//	coff.DataEntries[0].OffsetToData = offset - diroff
		case "/Relocations":
			coff.SectionHeader32.PointerToRelocations = offset
			coff.SectionHeader32.SizeOfRawData = offset - diroff
		case "/Symbols":
			coff.FileHeader.PointerToSymbolTable = offset
		}
		m := matcher{}
		switch {
		case m.Find(path, dir_n):
			coff.Dir.DirEntries[m[0]].OffsetToData = MASK_SUBDIRECTORY | (offset - diroff)
		case m.Find(path, dir_n_n):
			coff.Dir.Dirs[m[0]].DirEntries[m[1]].OffsetToData = MASK_SUBDIRECTORY | (offset - diroff)
		case m.Find(path, dataentry_n):
			direntry := <-leafwalker
			direntry.OffsetToData = offset - diroff
		case m.Find(path, dataentry_n_off):
			coff.Relocations[m[0]].RVA = offset - diroff
		case m.Find(path, data_n):
			coff.DataEntries[m[0]].OffsetToData = offset - diroff
		}

		if Plain(v.Kind()) {
			offset += uint32(binary.Size(v.Interface())) // TODO: change to v.Type().Size() ?
			return nil
		}
		vv, ok := v.Interface().(SizedReader)
		if ok {
			offset += uint32(vv.Size())
			return WALK_SKIP
		}
		return nil
	})

	// write the resulting file to disk
	Walk(coff, func(v reflect.Value, path string) error {
		if Plain(v.Kind()) {
			w.WriteLE(v.Interface())
			return nil
		}
		vv, ok := v.Interface().(SizedReader)
		if ok {
			w.WriteFromSized(vv)
			return WALK_SKIP
		}
		return nil
	})

	if w.Err != nil {
		return fmt.Errorf("Error writing output file: %s", w.Err)
	}

	return nil
}

func Plain(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}

func MustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

type matcher []int

func (m *matcher) Find(s string, re *regexp.Regexp) bool {
	subs := re.FindStringSubmatch(s)
	if subs == nil {
		return false
	}

	*m = (*m)[:0]
	for i := 1; i < len(subs); i++ {
		*m = append(*m, MustAtoi(subs[i]))
	}
	return true
}
