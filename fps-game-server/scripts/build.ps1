mkdir -Force ..\build

robocopy  /e ..\static ..\build\static

$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
go build -o ..\build\app-windows.exe ..\src\main.go

# $env:GOOS = 'linux'
# $env:GOARCH = 'amd64'
# go build -o ..\build\app-linux ..\src\main.go