#!/bin/bash

env GOOS=linux GOARCH=amd64 go build -o target/vpn_linux_amd64 main.go