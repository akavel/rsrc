rsrc - Tool for embedding Windows executable manifests in Go programs.

INSTALL: go get github.com/akavel/rsrc

USAGE: rsrc.exe -manifest FILE.exe.manifest [-o FILE.syso]
Generates a .syso file with specified resources embedded in .rsrc section,
aimed for consumption by Go linker when building Win32 excecutables.
OPTIONS:
  -manifest="": path to a Windows manifest file to embed
  -o="rsrc.syso": name of output COFF (.res or .syso) file

Just drop the generated *.syso file in the same directory with your *.go source files
of a Windows application, and it should link automatically when using `go build`.
You should not need to distribute a separate *.exe.manifest file with your GUI app
any more.

Based on ideas presented by Minux.

In case anything does not work, it'd be nice if you could report (either via Github
issues, or via email to czapkofan@gmail.com), and please attach the manifest file
which resulted in a problem, plus error message & symptoms, and/or any other details.

TODO:
- extend to allow embedding .ico icon files
- fix or remove FIXMEs

MAYBE LATER:
- extend to allow embedding arbitrary binary files
- extend to allow embedding binary files as linkable Go symbols
  (see http://code.google.com/p/go-wiki/wiki/GcToolchainTricks)
 - e.g. embed a binary in FILE.syso using GNU as(1)'s .incbin, then try to recreate similar result with rsrc

LICENSE: MIT
  Copyright 2013 Mateusz Czapliński <czapkofan@gmail.com>

http://github.com/akavel/rsrc
