rsrc - Tool for embedding binary resources in Go programs.

INSTALL: go get github.com/akavel/rsrc

USAGE:

rsrc -manifest FILE.exe.manifest [-ico FILE.ico] [-o FILE.syso]
  Generates a .syso file with specified resources embedded in .rsrc section,
  aimed for consumption by Go linker when building Win32 excecutables.

rsrc -data FILE.dat -o FILE.syso > FILE.c
  Generates a .syso file with specified opaque binary blob embedded,
  together with related .c file making it possible to access from Go code.
  Theoretically cross-platform, but reportedly cannot compile together with cgo.

The generated *.syso and *.c files should get automatically recognized
by 'go build' command and linked into an executable/library, as long as
there are any *.go files in the same directory.

OPTIONS:
  -data="": path to raw data file to embed
  -ico="": path to .ico file to embed , if you want to add more icons ,use a comma(,) to seperate the files ,such as a.ico,b.ico
  -manifest="": path to a Windows manifest file to embed
  -o="rsrc.syso": name of output COFF (.res or .syso) file

OUTPUTS:
This tool would print the ID you added to the resource file . This is a example outputs :
ID :   1 ,Add to resource , file : t005.exe.manifest
ID :   2 ,Add to resource , file : a.ico.0
ID :   3 ,Add to resource , file : a.ico
ID :   4 ,Add to resource , file : x.ico.0
ID :   5 ,Add to resource , file : x.ico


Based on ideas presented by Minux.

In case anything does not work, it'd be nice if you could report (either via Github
issues, or via email to czapkofan@gmail.com), and please attach the input file(s)
which resulted in a problem, plus error message & symptoms, and/or any other details.

TODO MAYBE/LATER:
- fix or remove FIXMEs

LICENSE: MIT
  Copyright 2013 Mateusz Czapliński <czapkofan@gmail.com>

http://github.com/akavel/rsrc
