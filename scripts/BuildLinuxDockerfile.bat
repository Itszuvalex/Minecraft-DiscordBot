@echo off
setlocal
set CURDIR=%~dp0
pushd "%CURDIR%"
pushd ..
@echo on
docker build --build-arg ExePath="%CD%" --build-arg ExeNameBuilt=%1 --build-arg ExeName=%2 -f build\Dockerfile .
@echo off
popd
popd
endlocal
