SHELL = bash

go_generate:
	go generate ./...

build:
	cd build/cmd
	go run .