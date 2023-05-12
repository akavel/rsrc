# rsrc

Tool for embedding binary resources in Go programs.

This fork is inspired by:

* [Original](https://github.com/akavel/rsrc) from `akavel`
* [Print Source with IDs](https://github.com/gonutz/rsrc/tree/main) from `gonutz`
* [Version Info](https://github.com/josephspurrier/goversioninfo) from `josephspurrier`
* [Version Info Variant](https://github.com/Doctible/rsrc) from `Doctible`

## Install

`go install github.com/dentalwings/rsrc`

## Usage

`rsrc.exe [-manifest <manifest>] [-ico FILE.ico[,FILE2.ico...]] [OPTIONS...]`

Generates a .syso file with specified resources embedded in .rsrc section,
aimed for consumption by Go linker when building Win32 excecutables.

The generated `*.syso` files should get automatically recognized by `go build`
command and linked into an executable/library, as long as there are any `*.go`
files in the same directory.

## Options

* `-arch <arch>`: Architecture of output file, one of: 386, amd64, arm
  (experimental), arm64 (experimental) (default "amd64")
* `-ico <icon>`: Comma-separated list of paths to .ico files to embed
* `-manifest <manifest>`: Path to a Windows manifest file to embed
* `-version <version>`: Path to a JSON file for version info
* `-o <output>`: Name of output COFF (.res or .syso) file; if set to empty,
  will default to `rsrc_windows_{arch}.syso`

Based on ideas presented by Minux.

In case anything does not work, it'd be nice if you could report (either via
Github issues, or via email to `czapkofan@gmail.com`), and please attach the
input file(s) which resulted in a problem, plus error message & symptoms,
and/or any other details.

## To Do, Maybe Later

* fix or remove FIXMEs

## License

MIT, Copyright 2013-2023 The rsrc Authors.

(NOTE: This project is currently in low-priority maintenance mode. I welcome
bug reports and sometimes try to address them, but this happens very randomly.
If it works for you, that's great and I'm happy! Still, and especially if not,
you might like to check the following more featureful alternative from @tc-hib
who is a very nice guy: [winres](https://github.com/tc-hib/go-winres))
