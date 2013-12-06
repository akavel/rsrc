package main

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unsafe"
)

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

	coffhdr := pe.FileHeader{
		Machine:              0x014c, //FIXME: find out how to differentiate this value, or maybe not necessary for Go
		NumberOfSections:     1,      // .rsrc
		TimeDateStamp:        0,      // was also 0 in sample data from MinGW's windres.exe
		PointerToSymbolTable: 0,
		NumberOfSymbols:      0,
		SizeOfOptionalHeader: 0,
		Characteristics:      0x0104, //FIXME: copied from windres.exe output, find out what should be here and why
	}
	err = binary.Write(out, binary.LittleEndian, coffhdr)
	if err != nil {
		return fmt.Errorf("Error writing COFF header: %s", err)
	}

	secthdr := pe.SectionHeader32{
		Name:             [8]byte{'.', 'r', 's', 'r', 'c', 0, 0, 0},
		SizeOfRawData:    uint32(len(manifest)), //FIXME: probably must include all the .rsrc directory structures too
		PointerToRawData: uint32(unsafe.Sizeof(pe.FileHeader{}) + unsafe.Sizeof(pe.SectionHeader32{})),
		Characteristics:  0x40000040, // "INITIALIZED_DATA MEM_READ" ?
	}
	err = binary.Write(out, binary.LittleEndian, secthdr)
	if err != nil {
		return fmt.Errorf("Error writing .rsrc section header: %s", err)
	}

	fmt.Println(string(manifest))
	fmt.Println(secthdr)

	return nil
}
