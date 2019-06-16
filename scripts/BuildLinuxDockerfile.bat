@echo off
setlocal
set CURDIR=%~dp0
pushd "%CURDIR%"
pushd ..
@echo on
docker build -f build\Dockerfile .
@echo off
popd
popd
endlocal
