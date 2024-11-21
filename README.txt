rsrc - Tool for embedding binary resources in Go programs.

INSTALL: go install github.com/akavel/rsrc

USAGE:

rsrc.exe [-manifest FILE.exe.manifest] [-ico FILE.ico[,FILE2.ico...]] [OPTIONS...]
  Generates a .syso file with specified resources embedded in .rsrc section,
  aimed for consumption by Go linker when building Win32 excecutables.

The generated *.syso files should get automatically recognized by 'go build'
command and linked into an executable/library, as long as there are any *.go
files in the same directory.

OPTIONS:
  -arch string
    	architecture of output file - one of: 386, amd64, [EXPERIMENTAL: arm, arm64] (default "amd64")
  -ico string
    	comma-separated list of paths to .ico files to embed
  -manifest string
    	path to a Windows manifest file to embed
  -o string
    	name of output COFF (.res or .syso) file; if set to empty, will default to 'rsrc_windows_{arch}.syso'

Based on ideas presented by Minux.

In case anything does not work, it'd be nice if you could report (either via Github
issues, or via email to czapkofan@gmail.com), and please attach the input file(s)
which resulted in a problem, plus error message & symptoms, and/or any other details.

TODO MAYBE/LATER:
- fix or remove FIXMEs

LICENSE: MIT
  Copyright 2013-2020 The rsrc Authors.

http://github.com/akavel/rsrc

(NOTE: This project is currently in low-priority maintenance mode. I welcome
 bug reports and sometimes try to address them, but this happens very randomly.
 If it works for you, that's great and I'm happy! Still, and especially if not,
 you might like to check the following more featureful alternative from @tc-hib
 who is a very nice guy:
 https://github.com/tc-hib/go-winres)

