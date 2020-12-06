package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const name = "rsrc.syso"

func TestIconBuildSucceeds(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(dir, "testdata")

	// Compile icon in testdata/ dir
	os.Stdout.Write([]byte("-- compiling icon...\n"))
	defer os.Remove(filepath.Join(dir, name))
	cmd := exec.Command("go", "run", "../rsrc.go", "-ico", "akavel.ico", "-arch", "amd64", "-o", name)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
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

	// TODO: test that we can extract icon from compiled app, and that it
	// is our icon
}

func TestManifestBuildSucceeds(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(dir, "testdata")

	// Compile icon in testdata/ dir
	os.Stdout.Write([]byte("-- compiling manifest...\n"))
	defer os.Remove(filepath.Join(dir, name))
	cmd := exec.Command("go", "run", "../rsrc.go", "-manifest", "manifest.xml", "-arch", "amd64", "-o", name)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
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

	// TODO: test that we can extract manifest from compiled app, and that it
	// is our manifest
}
