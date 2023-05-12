package rsrc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dentalwings/rsrc/binutil"
	"github.com/dentalwings/rsrc/coff"
	"github.com/dentalwings/rsrc/ico"
	"github.com/dentalwings/rsrc/internal"
	"github.com/josephspurrier/goversioninfo"
)

// on storing icons, see: http://blogs.msdn.com/b/oldnewthing/archive/2012/07/20/10331787.aspx
type _GRPICONDIR struct {
	ico.ICONDIR
	Entries []_GRPICONDIRENTRY
}

func (group _GRPICONDIR) Size() int64 {
	return int64(binary.Size(group.ICONDIR) + len(group.Entries)*binary.Size(group.Entries[0]))
}

type _GRPICONDIRENTRY struct {
	ico.IconDirEntryCommon
	Id uint16
}

func Embed(fnameout, arch, fnamein, fnameico, fnameversion string) (map[string]uint16, error) {
	ids := make(map[string]uint16)

	lastid := uint16(0)
	newid := func() uint16 {
		lastid++
		return lastid
	}

	out := coff.NewRSRC()
	err := out.Arch(arch)
	if err != nil {
		return nil, fmt.Errorf("setting arch to %v: %w", arch, err)
	}

	if fnamein != "" {
		manifest, err := binutil.SizedOpen(fnamein)
		if err != nil {
			return nil, fmt.Errorf("opening manifest file %v: %w", fnamein, err)
		}
		defer manifest.Close()

		id := newid()
		out.AddResource(coff.RT_MANIFEST, id, manifest)
		ids[fnamein] = id
	}

	if fnameico != "" {
		for _, fnameicosingle := range strings.Split(fnameico, ",") {
			f, iconId, err := addIcon(out, fnameicosingle, newid)
			if err != nil {
				return nil, fmt.Errorf("adding icon %v: %w", fnameicosingle, err)
			}
			defer f.Close()
			ids[fnameicosingle] = iconId
		}
	}

	if fnameversion != "" {
		if err := addVersion(out, fnameversion, newid); err != nil {
			return nil, fmt.Errorf("adding version info %s: %w", fnameversion, err)
		}
	}

	out.Freeze()
	return ids, internal.Write(out, fnameout)
}

func addIcon(out *coff.Coff, fname string, newid func() uint16) (io.Closer, uint16, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, 0, err
	}

	icons, err := ico.DecodeHeaders(f)
	if err != nil {
		f.Close()
		return nil, 0, err
	}

	gid := newid()
	if len(icons) > 0 {
		// RT_ICONs
		group := _GRPICONDIR{ICONDIR: ico.ICONDIR{
			Reserved: 0, // magic num.
			Type:     1, // magic num.
			Count:    uint16(len(icons)),
		}}
		for _, icon := range icons {
			id := newid()
			r := io.NewSectionReader(f, int64(icon.ImageOffset), int64(icon.BytesInRes))
			out.AddResource(coff.RT_ICON, id, r)
			group.Entries = append(group.Entries, _GRPICONDIRENTRY{icon.IconDirEntryCommon, id})
		}
		out.AddResource(coff.RT_GROUP_ICON, gid, group)
	}

	return f, gid, nil
}

func addVersion(out *coff.Coff, fname string, newid func() uint16) error {
	input, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("opening %v: %w", input, err)
	}
	defer input.Close()

	jsonBytes, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("reading %v: %w", input, err)
	}

	vi := &goversioninfo.VersionInfo{}
	if err := vi.ParseJSON(jsonBytes); err != nil {
		return fmt.Errorf("parsing JSON %v: %w", input, err)
	}

	vi.Build()
	vi.Walk()
	out.AddResource(coff.RT_VERSION, newid(), goversioninfo.SizedReader{Buffer: &vi.Buffer})

	return nil
}
