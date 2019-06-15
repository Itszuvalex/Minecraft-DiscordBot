@echo off
setlocal
set GOOS=linux
set GOARCH=amd64
set CURDIR=%~dp0
pushd "%CURDIR%"
pushd ..
@echo on
go build -o %1 cmd\mcdiscord\main.go
@echo off
popd
popd
endlocal
