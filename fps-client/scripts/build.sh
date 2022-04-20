#!/bin/bash

mkdir -p ../build

cp  -R ../static ../build

go build -o ../build/app-linux ../src/main.go
