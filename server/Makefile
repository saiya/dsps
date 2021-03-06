SHELL := /bin/bash

release_dir = dsps
go_module_files = go.mod go.sum
go_src_files = $(shell find . -type f -name '*.go')
build_deps = $(go_module_files) $(go_src_files) | $(release_dir)

DSPS_VERSION_ID ?= $(shell git rev-list -1 HEAD)
ldflags = "-X main.buildVersion=$(DSPS_VERSION_ID) -X main.buildAt=$(shell date +'%s')"

.PHONY: build test test.profile lint generate

build: $(release_dir).zip

test: generate
	GIN_MODE=release go test -v -race -timeout 30m -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

test.profile: generate
	mkdir -p pprof
	GIN_MODE=release go test -v -timeout 30m -bench -benchmem -o pprof/test.bin -cpuprofile pprof/cpu.out ./...
	go tool pprof --svg pprof/test.bin pprof/cpu.out > pprof/test.svg

lint: generate
# To run in local, authors recommends binary installation rather than module dependency: https://golangci-lint.run/usage/install/#local-installation
# On GitHub, authors recommends own GitHub Actions command: https://golangci-lint.run/usage/install/#ci-installation
	test -n "${CI}" || golangci-lint run ./...

# Check excess modules
	cp go.mod go.mod.bak
	go mod tidy
	diff go.mod go.mod.bak
	rm go.mod.bak

generate:
	go get github.com/golang/mock/mockgen
	go generate ./...

$(release_dir).zip: $(release_dir)/README.md $(release_dir)/CREDITS $(release_dir)/dsps-Linux-x86_64 $(release_dir)/dsps-Linux-aarch64 $(release_dir)/dsps-Windows-x86_64.exe $(release_dir)/dsps-Darwin-x86_64
	zip -r $@ $(release_dir)/ -x '$(release_dir)/.gitignore'

$(release_dir):
	mkdir $(release_dir)

$(release_dir)/CREDITS: $(go_module_files) | $(release_dir)
	go get github.com/Songmu/gocredits/cmd/gocredits
	rm -f $@
	gocredits -skip-missing . > $@

$(release_dir)/dsps-Linux-x86_64: $(build_deps)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ -ldflags "-X main.BuildDist=linux-amd64" -ldflags $(ldflags) main.go

$(release_dir)/dsps-Linux-aarch64: $(build_deps)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $@ -ldflags "-X main.BuildDist=linux-arm64" -ldflags $(ldflags) main.go

$(release_dir)/dsps-Windows-x86_64.exe: $(build_deps)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ -ldflags "-X main.BuildDist=windows-amd64" -ldflags $(ldflags) main.go

$(release_dir)/dsps-Darwin-x86_64: $(build_deps)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ -ldflags "-X main.BuildDist=darwin-amd64" -ldflags $(ldflags) main.go
