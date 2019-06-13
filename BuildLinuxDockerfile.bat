@echo off
set GOPATH=%~dp0
echo %GOPATH%
set GOOS=linux
set GOARCH=amd64
echo Building linux discord
go build -o mcdiscord.exe main
 echo Building Dockerfile
 docker build .
 echo Removing built linux exe
:: del mcdiscord.exe
:: echo Done!
