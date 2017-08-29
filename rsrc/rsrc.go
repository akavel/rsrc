package rsrc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/akavel/rsrc/binutil"
	"github.com/akavel/rsrc/coff"
	"github.com/akavel/rsrc/ico"
	"github.com/akavel/rsrc/internal"
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

func Embed(fnameout, arch, fnamein, fnameico string) error {
	newid := make(chan uint16)
	go func() {
		for i := uint16(1); ; i++ {
			newid <- i
		}
	}()

	coff := coff.NewRSRC()
	err := coff.Arch(arch)
	if err != nil {
		return err
	}

	if fnamein != "" {
		manifest, err := binutil.SizedOpen(fnamein)
		if err != nil {
			return fmt.Errorf("Error opening manifest file '%s': %s", fnamein, err)
		}
		defer manifest.Close()

		id := <-newid
		coff.AddResource(RT_MANIFEST, id, manifest)
		fmt.Println("Manifest ID: ", id)
	}
	if fnameico != "" {
		for _, fnameicosingle := range strings.Split(fnameico, ",") {
			err := addicon(coff, fnameicosingle, newid)
			if err != nil {
				return err
			}
		}
	}

	coff.Freeze()

	return internal.Write(coff, fnameout)
}

func addicon(coff *coff.Coff, fname string, newid <-chan uint16) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	//defer f.Close() don't defer, files will be closed by OS when app closes

	icons, err := ico.DecodeHeaders(f)
	if err != nil {
		return err
	}

	if len(icons) > 0 {
		// RT_ICONs
		group := GRPICONDIR{ICONDIR: ico.ICONDIR{
			Reserved: 0, // magic num.
			Type:     1, // magic num.
			Count:    uint16(len(icons)),
		}}
		for _, icon := range icons {
			id := <-newid
			r := io.NewSectionReader(f, int64(icon.ImageOffset), int64(icon.BytesInRes))
			coff.AddResource(RT_ICON, id, r)
			group.Entries = append(group.Entries, GRPICONDIRENTRY{icon.IconDirEntryCommon, id})
		}
		id := <-newid
		coff.AddResource(RT_GROUP_ICON, id, group)
		fmt.Println("Icon ", fname, " ID: ", id)
	}

	return nil
}
