package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const name = "rsrc_windows_amd64.syso"

func TestBuildSucceeds(t *testing.T) {
	tests := []struct {
		comment string
		args    []string
	}{{
		comment: "icon",
		args:    []string{"-ico", "akavel.ico"},
	}, {
		comment: "manifest",
		args:    []string{"-manifest", "manifest.xml"},
	}, {
		comment: "manifest & icon",
		args:    []string{"-manifest", "manifest.xml", "-ico", "akavel.ico"},
	}}
	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			dir = filepath.Join(dir, "testdata")

			// Compile icon/manifest in testdata/ dir
			os.Stdout.Write([]byte("-- compiling resource(s)...\n"))
			defer os.Remove(filepath.Join(dir, name))
			cmd := exec.Command("go", "run", "../rsrc.go", "-arch", "amd64")
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				t.Fatal(err)
			}

			// Verify if a .syso file with default name was created
			_, err = os.Stat(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}

			defer os.Setenv("GOOS", os.Getenv("GOOS"))
			defer os.Setenv("GOARCH", os.Getenv("GOARCH"))
			os.Setenv("GOOS", "windows")
			os.Setenv("GOARCH", "amd64")

			// Compile sample app in testdata/ dir, trying to link the icon
			// compiled above
			os.Stdout.Write([]byte("-- compiling app...\n"))
			cmd = exec.Command("go", "build")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				t.Fatal(err)
			}

			// Try running UPX on the executable, if the tool is found in PATH
			cmd = exec.Command("upx", "testdata.exe")
			if cmd.Path != "upx" {
				os.Stdout.Write([]byte("-- running upx...\n"))
				cmd.Dir = dir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				os.Stdout.Write([]byte("-- (upx not found)\n"))
			}

			// If on Windows, check that the compiled app can exec
			if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
				os.Stdout.Write([]byte("-- running app...\n"))
				cmd = &exec.Cmd{
					Path: "testdata.exe",
					Dir:  dir,
				}
				out, err := cmd.CombinedOutput()
				if err != nil {
					os.Stderr.Write(out)
					os.Stderr.Write([]byte("\n"))
					t.Fatal(err)
				}
				if string(out) != "hello world\n" {
					t.Fatalf("got unexpected output:\n%s", string(out))
				}
			}

			// TODO: test that we can extract icon/manifest from compiled app,
			// and that it is our icon/manifest
		})
	}
}
