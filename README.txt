USAGE: rsrc.exe -manifest FILE.exe.manifest [-o FILE.syso]
Generates a .syso file with specified resources embedded in .rsrc section,
aimed for consumption by Go linker when building Win32 excecutables.
OPTIONS:
  -manifest="": path to a Windows manifest file to embed
  -o="rsrc.syso": name of output COFF (.res or .syso) file