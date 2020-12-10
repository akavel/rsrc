@echo off
rem NOTE: see also:
rem https://github.com/golang/go/wiki/WindowsCrossCompiling
rem https://github.com/golang/go/wiki/InstallFromSource#install-c-tools
call :build rsrc windows_386
call :build rsrc windows_amd64
call :build rsrc linux_amd64
call :build rsrc darwin_amd64
set GOOS=
set GOARCH=
goto :eof

:build
set APP=%1
set PLATFORM=%2
:: Split param into GOOS & GOARCH (see: http://ss64.com/nt/syntax-substring.html)
set GOARCH=%PLATFORM:*_=%
call set GOOS=%%PLATFORM:_%GOARCH%=%%
:: Build filename
set FNAME=%APP%_%PLATFORM%
if "%GOOS%"=="windows" set FNAME=%FNAME%.exe
:: Do the build
echo == %FNAME% ==
go build -i -v -o %FNAME% .
goto :eof

