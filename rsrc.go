package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"

	"github.com/akavel/rsrc/binutil"
	"github.com/akavel/rsrc/coff"
	"github.com/akavel/rsrc/ico"
)

const (
	RT_ICON       = coff.RT_ICON
	RT_GROUP_ICON = coff.RT_GROUP_ICON
	RT_MANIFEST   = coff.RT_MANIFEST
)

// on storing icons, see: http://blogs.msdn.com/b/oldnewthing/archive/2012/07/20/10331787.aspx
type GRPICONDIR struct {
	ico.ICONDIR
	Entries []GRPICONDIRENTRY
}

func (group GRPICONDIR) Size() int64 {
	return int64(binary.Size(group.ICONDIR) + len(group.Entries)*binary.Size(group.Entries[0]))
}

type GRPICONDIRENTRY struct {
	ico.IconDirEntryCommon
	Id uint16
}

func main() {
	//TODO: allow in options advanced specification of multiple resources, as a tree (json?)
	//FIXME: verify that data file size doesn't exceed uint32 max value
	var fnamein, fnameico, fnamedata, fnameout string
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&fnamein, "manifest", "", "path to a Windows manifest file to embed")
	flags.StringVar(&fnameico, "ico", "", "path to .ico file to embed")
	flags.StringVar(&fnamedata, "data", "", "path to raw data file to embed")
	flags.StringVar(&fnameout, "o", "rsrc.syso", "name of output COFF (.res or .syso) file")
	_ = flags.Parse(os.Args[1:])
	if fnamein == "" && fnamedata == "" {
		fmt.Fprintf(os.Stderr, "USAGE: %s -manifest FILE.exe.manifest [-ico FILE.ico] [-o FILE.syso]\n"+
			"       %s -data FILE.dat [-o FILE.syso] > embed.c\n"+
			"Generates a .syso file with specified resources embedded in .rsrc section,\n"+
			"aimed for consumption by Go linker when building Win32 excecutables.\n"+
			"OPTIONS:\n",
			os.Args[0], os.Args[0])
		flags.PrintDefaults()
		os.Exit(1)
	}

	var err error
	switch {
	case fnamein != "":
		err = run(fnamein, fnameico, fnameout)
	case fnamedata != "":
		err = rundata(fnamedata, fnameout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rundata(fnamedata, fnameout string) error {
	dat, err := binutil.SizedOpen(fnamedata)
	if err != nil {
		return fmt.Errorf("Error opening data file '%s': %s", fnamedata, err)
	}
	defer dat.Close()

	symname := fmt.Sprintf("_rsrc_%04x", rand.Uint32())

	coff := coff.NewRDATA()
	coff.AddData(symname+"_begin", symname+"_end", dat)
	coff.Freeze()
	err = write(coff, fnameout)
	if err != nil {
		return err
	}

	//FIXME: output a .c file

	return nil
}

func run(fnamein, fnameico, fnameout string) error {
	manifest, err := binutil.SizedOpen(fnamein)
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

	coff := coff.NewRSRC()

	coff.AddResource(RT_MANIFEST, <-newid, manifest)

	if len(icons) > 0 {
		// RT_ICONs
		group := GRPICONDIR{ICONDIR: ico.ICONDIR{
			Reserved: 0, // magic num.
			Type:     1, // magic num.
			Count:    uint16(len(icons)),
		}}
		for _, icon := range icons {
			id := <-newid
			r := io.NewSectionReader(iconsf, int64(icon.ImageOffset), int64(icon.BytesInRes))

			coff.AddResource(RT_ICON, id, r)
			group.Entries = append(group.Entries, GRPICONDIRENTRY{icon.IconDirEntryCommon, id})
		}

		// RT_GROUP_ICON
		coff.AddResource(RT_GROUP_ICON, <-newid, group)
	}

	coff.Freeze()

	return write(coff, fnameout)
}

func write(coff *coff.Coff, fnameout string) error {
	out, err := os.Create(fnameout)
	if err != nil {
		return err
	}
	defer out.Close()
	w := binutil.Writer{W: out}

	// write the resulting file to disk
	binutil.Walk(coff, func(v reflect.Value, path string) error {
		if binutil.Plain(v.Kind()) {
			w.WriteLE(v.Interface())
			return nil
		}
		vv, ok := v.Interface().(binutil.SizedReader)
		if ok {
			w.WriteFromSized(vv)
			return binutil.WALK_SKIP
		}
		return nil
	})

	if w.Err != nil {
		return fmt.Errorf("Error writing output file: %s", w.Err)
	}

	return nil
}
