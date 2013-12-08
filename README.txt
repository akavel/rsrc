USAGE: rsrc.exe -manifest FILE.exe.manifest [-o FILE.syso]
Generates a .syso file with specified resources embedded in .rsrc section,
aimed for consumption by Go linker when building Win32 excecutables.
OPTIONS:
  -manifest="": path to a Windows manifest file to embed
  -o="rsrc.syso": name of output COFF (.res or .syso) file

TODO:
- extend to allow embedding .ico icon files
- fix or remove FIXMEs

MAYBE LATER:
- extend to allow embedding arbitrary binary files
- extend to allow embedding binary files as linkable Go symbols (see http://code.google.com/p/go-wiki/wiki/GcToolchainTricks)
 - e.g. embed a binary in FILE.syso using GNU as(1)'s .incbin, then try to recreate similar result with rsrc
